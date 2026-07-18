package grpc

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/harmost/agent/internal/docker"
	"github.com/harmost/agent/internal/metrics"
	harmostv1 "github.com/harmost/proto/gen/harmost/v1"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	statusBuffer = 256
	logBuffer    = 1024
)

type Client struct {
	mgr *docker.Manager // nil when Docker is unavailable on this host

	// statusCh and logCh live for the process, not one Connect call, so
	// messages queued while the stream is down survive a reconnect. Jobs
	// capture Send once at dispatch and it stays valid forever.
	statusCh chan *harmostv1.AgentMessage
	logCh    chan *harmostv1.AgentMessage

	// pendingStatus is a StatusUpdate whose stream.Send failed; it is resent
	// first on the next connection. Only the Connect goroutine touches it.
	pendingStatus *harmostv1.AgentMessage
}

func New(mgr *docker.Manager) *Client {
	return &Client{
		mgr:      mgr,
		statusCh: make(chan *harmostv1.AgentMessage, statusBuffer),
		logCh:    make(chan *harmostv1.AgentMessage, logBuffer),
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

func (c *Client) Connect(ctx context.Context, target, token string) error {
	conn, err := googlegrpc.NewClient(target, googlegrpc.WithTransportCredentials(insecure.NewCredentials()))
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
			c.handleMessage(ctx, msg)
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
			return err
		case msg := <-c.statusCh:
			if err := sendMsg(msg); err != nil {
				return err
			}
		case msg := <-c.logCh:
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
// reconnect.

func (c *Client) handleMessage(ctx context.Context, msg *harmostv1.HubMessage) {
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
	}
}

func Target(hubURL string) string {
	for _, prefix := range []string{"https://", "http://"} {
		if len(hubURL) > len(prefix) && hubURL[:len(prefix)] == prefix {
			return hubURL[len(prefix):]
		}
	}
	return hubURL
}
