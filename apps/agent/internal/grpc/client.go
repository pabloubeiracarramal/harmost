package grpc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/harmost/agent/internal/config"
	"github.com/harmost/agent/internal/docker"
	"github.com/harmost/agent/internal/metrics"
	harmostv1 "github.com/harmost/proto/gen/harmost/v1"
	"github.com/moby/moby/api/types/container"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	statusBuffer = 256
	logBuffer    = 1024

	// containerWatchInterval is independent of the 30s heartbeat — the
	// containers card wants something closer to live.
	containerWatchInterval = 5 * time.Second
)

// ErrUnpaired is returned by Connect when the stream ended because the hub
// sent an Unpair message — the daemon loop should stop reconnecting rather
// than retry with a now-revoked token.
var ErrUnpaired = errors.New("agent unpaired from hub")

type Client struct {
	mgr *docker.Manager // nil when Docker is unavailable on this host

	// statusCh and logCh live for the process, not one Connect call, so
	// messages queued while the stream is down survive a reconnect. Jobs
	// capture Send once at dispatch and it stays valid forever.
	statusCh chan *harmostv1.AgentMessage
	logCh    chan *harmostv1.AgentMessage
	// containerCh holds at most one pending ContainerListUpdate: only the
	// latest list is ever worth sending, so Send coalesces instead of
	// queuing (unlike statusCh/logCh).
	containerCh chan *harmostv1.AgentMessage

	// pendingStatus is a StatusUpdate whose stream.Send failed; it is resent
	// first on the next connection. Only the Connect goroutine touches it.
	pendingStatus *harmostv1.AgentMessage

	// watchMu guards watchCancel, the currently running container push
	// loop (if any). Scoped to one Connect call — see startWatchingContainers.
	watchMu     sync.Mutex
	watchCancel context.CancelFunc

	// unpaired is set on the recv goroutine when a HubMessage_Unpair
	// arrives, and read by Connect's main loop once the stream then closes,
	// to distinguish an unpair from an ordinary disconnect.
	unpaired atomic.Bool
}

func New(mgr *docker.Manager) *Client {
	return &Client{
		mgr:         mgr,
		statusCh:    make(chan *harmostv1.AgentMessage, statusBuffer),
		logCh:       make(chan *harmostv1.AgentMessage, logBuffer),
		containerCh: make(chan *harmostv1.AgentMessage, 1),
	}
}

// Send queues a message for delivery to the hub; it satisfies docker.SendFunc
// and never blocks. Log chunks are best-effort and dropped when their buffer
// is full. A terminal StatusUpdate evicts the oldest queued message instead of
// being dropped — losing it would make the hub's reconcile mislabel a finished
// job as lost.
func (c *Client) Send(msg *harmostv1.AgentMessage) {
	if _, ok := msg.Payload.(*harmostv1.AgentMessage_LogChunk); ok {
		select {
		case c.logCh <- msg:
		default:
			log.Printf("agent: log buffer full, dropping line")
		}
		return
	}

	if _, ok := msg.Payload.(*harmostv1.AgentMessage_ContainerList); ok {
		select {
		case <-c.containerCh: // drop the stale pending list, if any
		default:
		}
		select {
		case c.containerCh <- msg:
		default:
		}
		return
	}

	for {
		select {
		case c.statusCh <- msg:
			return
		default:
		}
		if !isTerminalStatus(msg) {
			log.Printf("agent: status buffer full, dropping message")
			return
		}
		select {
		case <-c.statusCh:
			log.Printf("agent: status buffer full, evicted oldest message for terminal status")
		default:
		}
	}
}

func isTerminalStatus(msg *harmostv1.AgentMessage) bool {
	su, ok := msg.Payload.(*harmostv1.AgentMessage_StatusUpdate)
	if !ok {
		return false
	}
	switch su.StatusUpdate.State {
	case harmostv1.JobState_JOB_STATE_SUCCEEDED,
		harmostv1.JobState_JOB_STATE_FAILED,
		harmostv1.JobState_JOB_STATE_CANCELLED,
		harmostv1.JobState_JOB_STATE_TIMED_OUT:
		return true
	}
	return false
}

// Connect dials the hub and holds the bidi stream open. TLS via the system
// cert pool is used unless insecureConn is set, which is for local dev only
// (hub serves plaintext gRPC unless GRPC_TLS_CERT_FILE/GRPC_TLS_KEY_FILE are
// configured — see docs/dev.md).
func (c *Client) Connect(ctx context.Context, target, token string, insecureConn bool) error {
	c.unpaired.Store(false)

	// A watch loop left over from a previous stream (if the hub never got to
	// send Unwatch before the connection dropped) is stale — its own ctx is
	// already cancelled below on that call's return, but reset the field so
	// a fresh Watch on this connection isn't mistaken for "already running".
	c.stopWatchingContainers()

	// watchStreamCtx bounds anything that must not outlive this particular
	// stream (the container watch loop) — unlike ctx, the daemon's root
	// context, which jobs bind to so they survive a reconnect.
	watchStreamCtx, cancelWatchStream := context.WithCancel(ctx)
	defer cancelWatchStream()

	creds := credentials.NewClientTLSFromCert(nil, "")
	if insecureConn {
		creds = insecure.NewCredentials()
	}
	conn, err := googlegrpc.NewClient(target, googlegrpc.WithTransportCredentials(creds))
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	client := harmostv1.NewAgentServiceClient(conn)

	md := metadata.Pairs("authorization", "Bearer "+token)
	streamCtx := metadata.NewOutgoingContext(ctx, md)

	stream, err := client.Connect(streamCtx)
	if err != nil {
		return fmt.Errorf("open stream: %w", err)
	}

	var runningIDs []string
	if c.mgr != nil {
		runningIDs = c.mgr.RunningJobIDs()
	}

	hostname, _ := os.Hostname()
	if err := stream.Send(&harmostv1.AgentMessage{
		Payload: &harmostv1.AgentMessage_Hello{
			Hello: &harmostv1.AgentHello{
				Name:          hostname,
				Description:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
				Version:       "0.1.0",
				Hostname:      hostname,
				RunningJobIds: runningIDs,
			},
		},
	}); err != nil {
		return fmt.Errorf("send hello: %w", err)
	}

	log.Printf("agent: connected to hub")

	// gRPC streams allow only one sending goroutine, so the select loop
	// below is the sole caller of stream.Send; everything else funnels
	// through the process-lifetime channels via Send.
	sendMsg := func(msg *harmostv1.AgentMessage) error {
		if err := stream.Send(msg); err != nil {
			if _, ok := msg.Payload.(*harmostv1.AgentMessage_StatusUpdate); ok {
				c.pendingStatus = msg
			}
			return fmt.Errorf("send: %w", err)
		}
		return nil
	}

	if c.pendingStatus != nil {
		msg := c.pendingStatus
		c.pendingStatus = nil
		if err := sendMsg(msg); err != nil {
			return err
		}
	}

	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	recvCh := make(chan error, 1)
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				recvCh <- err
				return
			}
			c.handleMessage(ctx, watchStreamCtx, msg)
		}
	}()

	for {
		// Statuses drain ahead of logs so a flood of log lines can't
		// delay a state transition.
		select {
		case msg := <-c.statusCh:
			if err := sendMsg(msg); err != nil {
				return err
			}
			continue
		default:
		}

		select {
		case <-ctx.Done():
			return nil
		case err := <-recvCh:
			if c.unpaired.Load() {
				return ErrUnpaired
			}
			return err
		case msg := <-c.statusCh:
			if err := sendMsg(msg); err != nil {
				return err
			}
		case msg := <-c.logCh:
			if err := sendMsg(msg); err != nil {
				return err
			}
		case msg := <-c.containerCh:
			if err := sendMsg(msg); err != nil {
				return err
			}
		case t := <-heartbeatTicker.C:
			if err := stream.Send(&harmostv1.AgentMessage{
				Payload: &harmostv1.AgentMessage_Heartbeat{
					Heartbeat: &harmostv1.Heartbeat{
						Timestamp: timestamppb.New(t),
						Metrics:   metrics.Collect(ctx),
					},
				},
			}); err != nil {
				return fmt.Errorf("send heartbeat: %w", err)
			}
		}
	}
}

// handleMessage runs on the recv goroutine; anything long-lived must be
// dispatched to its own goroutine (the manager does this) so a slow job
// can't stall hub messages. ctx is the daemon's root context — jobs are
// bound to the daemon's lifetime, not the stream's, so they survive a
// reconnect. watchStreamCtx bounds the container watch loop instead: it
// must stop when this stream drops, not persist across a reconnect.
func (c *Client) handleMessage(ctx, watchStreamCtx context.Context, msg *harmostv1.HubMessage) {
	switch payload := msg.Payload.(type) {
	case *harmostv1.HubMessage_DispatchJob:
		log.Printf("agent: received job %s", payload.DispatchJob.JobId)
		if c.mgr == nil {
			c.Send(&harmostv1.AgentMessage{Payload: &harmostv1.AgentMessage_StatusUpdate{
				StatusUpdate: &harmostv1.JobStatusUpdate{
					JobId:     payload.DispatchJob.JobId,
					State:     harmostv1.JobState_JOB_STATE_FAILED,
					Timestamp: timestamppb.Now(),
					Message:   "docker is not available on this agent",
				},
			}})
			return
		}
		c.mgr.Dispatch(ctx, payload.DispatchJob.JobId, payload.DispatchJob.Spec, c.Send)
	case *harmostv1.HubMessage_CancelJob:
		if c.mgr == nil || !c.mgr.Cancel(payload.CancelJob.JobId) {
			log.Printf("agent: cancel for unknown job %s", payload.CancelJob.JobId)
		}
	case *harmostv1.HubMessage_Ping:
		c.Send(&harmostv1.AgentMessage{Payload: &harmostv1.AgentMessage_Pong{
			Pong: &harmostv1.Pong{
				PingSentAt: payload.Ping.SentAt,
				ReceivedAt: timestamppb.Now(),
			},
		}})
	case *harmostv1.HubMessage_WatchContainers:
		c.startWatchingContainers(watchStreamCtx)
	case *harmostv1.HubMessage_UnwatchContainers:
		c.stopWatchingContainers()
	case *harmostv1.HubMessage_ContainerAction:
		// Stop/Restart can take several seconds (Docker's default 10s grace
		// before SIGKILL) — must not block this goroutine, which also
		// drives every other incoming message.
		go c.performContainerAction(ctx, payload.ContainerAction)
	case *harmostv1.HubMessage_Unpair:
		log.Printf("agent: unpaired from hub")
		if err := config.Clear(); err != nil {
			log.Printf("agent: failed to clear local pairing config: %v", err)
		}
		c.unpaired.Store(true)
	}
}

// performContainerAction runs one lifecycle action, reports the outcome,
// then immediately refreshes the container list — belt-and-suspenders on
// top of the ~5s watch tick so the UI doesn't sit waiting for it.
func (c *Client) performContainerAction(ctx context.Context, req *harmostv1.ContainerActionRequest) {
	var err error
	switch {
	case c.mgr == nil:
		err = fmt.Errorf("docker is not available on this agent")
	default:
		switch req.Action {
		case harmostv1.ContainerAction_CONTAINER_ACTION_START:
			err = c.mgr.StartContainer(ctx, req.ContainerId)
		case harmostv1.ContainerAction_CONTAINER_ACTION_STOP:
			err = c.mgr.StopContainer(ctx, req.ContainerId)
		case harmostv1.ContainerAction_CONTAINER_ACTION_RESTART:
			err = c.mgr.RestartContainer(ctx, req.ContainerId)
		case harmostv1.ContainerAction_CONTAINER_ACTION_REMOVE:
			err = c.mgr.RemoveContainer(ctx, req.ContainerId)
		default:
			err = fmt.Errorf("unknown action %s", req.Action)
		}
	}

	result := &harmostv1.ContainerActionResult{
		ContainerId: req.ContainerId,
		Action:      req.Action,
		Success:     err == nil,
	}
	if err != nil {
		result.Error = err.Error()
	}
	c.Send(&harmostv1.AgentMessage{Payload: &harmostv1.AgentMessage_ContainerActionResult{
		ContainerActionResult: result,
	}})

	c.pushContainerList(ctx)
}

// startWatchingContainers begins pushing ContainerListUpdate every
// containerWatchInterval until stopWatchingContainers is called or ctx is
// done. A second WatchContainersRequest on top of an already-running loop
// (the hub may resend, e.g. on a watcher-count race) is a no-op.
func (c *Client) startWatchingContainers(ctx context.Context) {
	c.watchMu.Lock()
	defer c.watchMu.Unlock()
	if c.watchCancel != nil {
		return
	}
	watchCtx, cancel := context.WithCancel(ctx)
	c.watchCancel = cancel
	go c.watchContainersLoop(watchCtx)
}

func (c *Client) stopWatchingContainers() {
	c.watchMu.Lock()
	defer c.watchMu.Unlock()
	if c.watchCancel != nil {
		c.watchCancel()
		c.watchCancel = nil
	}
}

func (c *Client) watchContainersLoop(ctx context.Context) {
	c.pushContainerList(ctx)
	if c.mgr == nil {
		// No Docker on this host: the one empty list above is all there
		// ever will be, so there's nothing to poll on a ticker.
		return
	}

	ticker := time.NewTicker(containerWatchInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.pushContainerList(ctx)
		}
	}
}

// pushContainerList lists every container on the host, any state, and
// attaches a live resource-usage sample to the ones currently running (the
// only state the stats endpoint accepts). One bad container's stats call
// logs and is skipped rather than blanking the whole tick.
func (c *Client) pushContainerList(ctx context.Context) {
	var infos []*harmostv1.ContainerInfo
	if c.mgr != nil {
		list, err := c.mgr.ListContainers(ctx)
		if err != nil {
			log.Printf("agent: list containers: %v", err)
		} else {
			infos = make([]*harmostv1.ContainerInfo, len(list))
			for i, item := range list {
				info := toContainerInfo(item)
				if info.State == string(container.StateRunning) {
					if cpuPct, memUsage, memLimit, err := c.mgr.ContainerStats(ctx, info.Id); err != nil {
						log.Printf("agent: stats for %s: %v", info.Id, err)
					} else {
						info.Stats = &harmostv1.ContainerStats{
							CpuUsagePercent:  float32(cpuPct),
							MemoryUsageBytes: memUsage,
							MemoryLimitBytes: memLimit,
						}
					}
				}
				infos[i] = info
			}
		}
	}
	c.Send(&harmostv1.AgentMessage{Payload: &harmostv1.AgentMessage_ContainerList{
		ContainerList: &harmostv1.ContainerListUpdate{Containers: infos},
	}})
}

func toContainerInfo(c container.Summary) *harmostv1.ContainerInfo {
	name := c.ID
	if len(c.Names) > 0 {
		name = strings.TrimPrefix(c.Names[0], "/")
	}
	return &harmostv1.ContainerInfo{
		Id:        c.ID,
		Image:     c.Image,
		Name:      name,
		State:     string(c.State),
		Status:    c.Status,
		StartedAt: timestamppb.New(time.Unix(c.Created, 0)),
		Ports:     toContainerPorts(c.Ports),
		Volumes:   toContainerMounts(c.Mounts),
	}
}

func toContainerPorts(ports []container.PortSummary) []*harmostv1.ContainerPort {
	out := make([]*harmostv1.ContainerPort, len(ports))
	for i, p := range ports {
		var hostIP string
		if p.IP.IsValid() {
			hostIP = p.IP.String()
		}
		out[i] = &harmostv1.ContainerPort{
			HostIp:      hostIP,
			PrivatePort: uint32(p.PrivatePort),
			PublicPort:  uint32(p.PublicPort),
			Type:        p.Type,
		}
	}
	return out
}

func toContainerMounts(mounts []container.MountPoint) []*harmostv1.ContainerMount {
	out := make([]*harmostv1.ContainerMount, len(mounts))
	for i, m := range mounts {
		out[i] = &harmostv1.ContainerMount{
			Type:        string(m.Type),
			Name:        m.Name,
			Source:      m.Source,
			Destination: m.Destination,
			// RW means writable (read-write); the wire field is the negation.
			ReadOnly: !m.RW,
		}
	}
	return out
}

func Target(hubURL string) string {
	for _, prefix := range []string{"https://", "http://"} {
		if len(hubURL) > len(prefix) && hubURL[:len(prefix)] == prefix {
			return hubURL[len(prefix):]
		}
	}
	return hubURL
}
