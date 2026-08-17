package grpcapi

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/harmost/hub/internal/domain"
	"github.com/harmost/hub/internal/events"
	harmostv1 "github.com/harmost/proto/gen/harmost/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	grpcstatus "google.golang.org/grpc/status"
)

const (
	logFlushInterval = 500 * time.Millisecond
	logBatchSize     = 500

	// reconcileDelay is how long after a hello the hub waits before failing
	// jobs the agent did not report — long enough for terminal statuses
	// queued agent-side during an outage to land first.
	reconcileDelay = 15 * time.Second
)

// Connect is the bidirectional stream handler. Each connected agent gets one
// long-lived call to this method.
func (s *Server) Connect(stream grpc.BidiStreamingServer[harmostv1.AgentMessage, harmostv1.HubMessage]) error {
	ctx := stream.Context()

	orgID, agentID, err := orgIDFromToken(ctx, s)
	if err != nil {
		return err
	}

	// First message must be AgentHello.
	first, err := stream.Recv()
	if err != nil {
		return err
	}
	hello, ok := first.Payload.(*harmostv1.AgentMessage_Hello)
	if !ok {
		return grpcstatus.Error(codes.InvalidArgument, "first message must be AgentHello")
	}

	input := domain.AgentConnectInput{
		Name:        hello.Hello.Name,
		Description: hello.Hello.Description,
		Version:     hello.Hello.Version,
		Hostname:    hello.Hello.Hostname,
	}

	var agent *domain.Agent
	if agentID != "" {
		agent, err = s.svc.Agent.UpdateOnConnect(ctx, agentID, input)
	} else {
		agent, err = s.svc.Agent.Connect(ctx, orgID, input)
	}
	if err != nil {
		return grpcstatus.Errorf(codes.Internal, "agent registration: %v", err)
	}

	s.bus.Publish(events.Event{
		Type:    events.AgentConnected,
		OrgID:   orgID,
		AgentID: agent.ID,
		At:      time.Now(),
	})

	// Reconcile against the hello's running set once queued terminal
	// statuses have had a chance to land. context.Background(): the answer
	// is valid even if this stream has dropped again by then.
	helloAt := time.Now()
	runningIDs := hello.Hello.RunningJobIds
	time.AfterFunc(reconcileDelay, func() {
		if err := s.svc.Job.ReconcileAgent(context.Background(), agent.ID, runningIDs, helloAt); err != nil {
			log.Printf("reconcile agent %s: %v", agent.ID, err)
		}
	})

	var sendMu sync.Mutex
	safeSend := func(msg *harmostv1.HubMessage) error {
		sendMu.Lock()
		defer sendMu.Unlock()
		return stream.Send(msg)
	}

	killCh := s.reg.add(agent.ID, safeSend)
	defer s.reg.remove(agent.ID)

	// A watch from before this (re)connect doesn't survive on the agent
	// side — its own stream-scoped state died with the old stream — so
	// resend it if anyone is still watching.
	if s.reg.watcherCount(agent.ID) > 0 {
		_ = safeSend(&harmostv1.HubMessage{
			Payload: &harmostv1.HubMessage_WatchContainers{
				WatchContainers: &harmostv1.WatchContainersRequest{},
			},
		})
	}
	defer func() {
		s.svc.Agent.Disconnect(context.Background(), agent.ID)
		s.bus.Publish(events.Event{
			Type:    events.AgentDisconnected,
			OrgID:   orgID,
			AgentID: agent.ID,
			At:      time.Now(),
		})
	}()

	type recvResult struct {
		msg *harmostv1.AgentMessage
		err error
	}
	recvCh := make(chan recvResult, 32)
	go func() {
		for {
			msg, err := stream.Recv()
			recvCh <- recvResult{msg, err}
			if err != nil {
				return
			}
		}
	}()

	var logBuf []domain.JobLog
	ticker := time.NewTicker(logFlushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(logBuf) == 0 {
			return
		}
		if err := s.svc.JobLog.IngestChunks(context.Background(), logBuf); err == nil {
			// One job.log event per job per flush batch — never per line,
			// or a chatty job would flood the 64-slot subscriber buffers.
			byJob := make(map[string][]events.LogLine)
			for _, l := range logBuf {
				byJob[l.JobID] = append(byJob[l.JobID], events.LogLine{
					Line:      l.Line,
					Stream:    events.LogStream(l.Stream),
					Sequence:  l.Sequence,
					Timestamp: l.Timestamp,
				})
			}
			for jobID, lines := range byJob {
				s.bus.Publish(events.Event{
					Type:    events.JobLog,
					OrgID:   orgID,
					AgentID: agent.ID,
					JobID:   jobID,
					At:      time.Now(),
					Payload: events.JobLogPayload{Lines: lines},
				})
			}
		}
		logBuf = logBuf[:0]
	}

	for {
		select {
		case r := <-recvCh:
			if r.err != nil {
				flush()
				return r.err
			}
			s.handleMessage(ctx, agent.ID, orgID, r.msg, &logBuf, flush)

		case <-ticker.C:
			flush()

		case <-killCh:
			flush()
			return grpcstatus.Error(codes.Aborted, "agent unpaired")

		case <-ctx.Done():
			flush()
			return ctx.Err()
		}
	}
}

func (s *Server) handleMessage(
	ctx context.Context,
	agentID string,
	orgID string,
	msg *harmostv1.AgentMessage,
	logBuf *[]domain.JobLog,
	flush func(),
) {
	switch p := msg.Payload.(type) {
	case *harmostv1.AgentMessage_StatusUpdate:
		u := p.StatusUpdate
		in := domain.JobStatusInput{
			JobID:     u.JobId,
			State:     protoStateToJobState(u.State),
			Message:   u.Message,
			Timestamp: u.Timestamp.AsTime(),
			ExitCode:  u.ExitCode,
		}
		_ = s.svc.Job.HandleStatusUpdate(ctx, in)

	case *harmostv1.AgentMessage_LogChunk:
		*logBuf = append(*logBuf, protoChunkToJobLog(p.LogChunk))
		if len(*logBuf) >= logBatchSize {
			flush()
		}

	case *harmostv1.AgentMessage_Heartbeat:
		at := p.Heartbeat.Timestamp.AsTime()
		var m domain.AgentMetrics
		if p.Heartbeat.Metrics != nil {
			pm := p.Heartbeat.Metrics
			m = domain.AgentMetrics{
				CpuUsagePercent:   pm.CpuUsagePercent,
				MemoryUsedBytes:   pm.MemoryUsedBytes,
				MemoryTotalBytes:  pm.MemoryTotalBytes,
				DiskUsedBytes:     pm.DiskUsedBytes,
				DiskTotalBytes:    pm.DiskTotalBytes,
				RunningContainers: pm.RunningContainers,
			}
		}
		_ = s.svc.Agent.HandleHeartbeat(ctx, agentID, m, at)
		s.bus.Publish(events.Event{
			Type:    events.AgentHeartbeat,
			OrgID:   orgID,
			AgentID: agentID,
			At:      at,
		})

	case *harmostv1.AgentMessage_Pong:
		// latency tracking not implemented yet

	case *harmostv1.AgentMessage_ContainerList:
		containers := make([]events.ContainerInfo, len(p.ContainerList.Containers))
		for i, c := range p.ContainerList.Containers {
			containers[i] = protoContainerToEvent(c)
		}
		s.bus.Publish(events.Event{
			Type:    events.AgentContainers,
			OrgID:   orgID,
			AgentID: agentID,
			At:      time.Now(),
			Payload: events.ContainersPayload{Containers: containers},
		})

	case *harmostv1.AgentMessage_ContainerActionResult:
		r := p.ContainerActionResult
		s.bus.Publish(events.Event{
			Type:    events.AgentContainerAction,
			OrgID:   orgID,
			AgentID: agentID,
			At:      time.Now(),
			Payload: events.ContainerActionPayload{
				ContainerID: r.ContainerId,
				Action:      events.ContainerActionKind(protoActionToString[r.Action]),
				Success:     r.Success,
				Error:       r.Error,
			},
		})
	}
}

var protoActionToString = map[harmostv1.ContainerAction]string{
	harmostv1.ContainerAction_CONTAINER_ACTION_START:   "start",
	harmostv1.ContainerAction_CONTAINER_ACTION_STOP:    "stop",
	harmostv1.ContainerAction_CONTAINER_ACTION_RESTART: "restart",
	harmostv1.ContainerAction_CONTAINER_ACTION_REMOVE:  "remove",
}

func protoContainerToEvent(c *harmostv1.ContainerInfo) events.ContainerInfo {
	ports := make([]events.ContainerPort, len(c.Ports))
	for i, p := range c.Ports {
		ports[i] = events.ContainerPort{
			HostIP:      p.HostIp,
			PrivatePort: int32(p.PrivatePort),
			PublicPort:  int32(p.PublicPort),
			Type:        p.Type,
		}
	}
	volumes := make([]events.ContainerMount, len(c.Volumes))
	for i, v := range c.Volumes {
		volumes[i] = events.ContainerMount{
			Type:        v.Type,
			Name:        v.Name,
			Source:      v.Source,
			Destination: v.Destination,
			ReadOnly:    v.ReadOnly,
		}
	}
	var stats *events.ContainerStats
	if c.Stats != nil {
		stats = &events.ContainerStats{
			CPUUsagePercent:  c.Stats.CpuUsagePercent,
			MemoryUsageBytes: c.Stats.MemoryUsageBytes,
			MemoryLimitBytes: c.Stats.MemoryLimitBytes,
		}
	}
	return events.ContainerInfo{
		ID:        c.Id,
		Image:     c.Image,
		Name:      c.Name,
		State:     c.State,
		Status:    c.Status,
		StartedAt: c.StartedAt.AsTime(),
		Ports:     ports,
		Volumes:   volumes,
		Stats:     stats,
	}
}

// orgIDFromToken validates the agent token from gRPC metadata and returns the org ID and agent ID.
func orgIDFromToken(ctx context.Context, s *Server) (orgID, agentID string, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", "", grpcstatus.Error(codes.Unauthenticated, "no metadata")
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return "", "", grpcstatus.Error(codes.Unauthenticated, "missing authorization metadata")
	}
	token := strings.TrimPrefix(vals[0], "Bearer ")
	if token == "" {
		return "", "", grpcstatus.Error(codes.Unauthenticated, "empty token")
	}
	orgID, agentID, err = s.svc.AgentToken.Validate(ctx, token)
	if err != nil {
		return "", "", grpcstatus.Error(codes.Unauthenticated, "invalid agent token")
	}
	return orgID, agentID, nil
}
