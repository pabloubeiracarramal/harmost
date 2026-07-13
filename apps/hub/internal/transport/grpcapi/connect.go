package grpcapi

import (
	"context"
	"strings"
	"sync"
	"time"

	harmostv1 "github.com/harmost/proto/gen/harmost/v1"
	"github.com/harmost/hub/internal/domain"
	"github.com/harmost/hub/internal/events"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	grpcstatus "google.golang.org/grpc/status"
)

const (
	logFlushInterval = 500 * time.Millisecond
	logBatchSize     = 500
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

	var sendMu sync.Mutex
	safeSend := func(msg *harmostv1.HubMessage) error {
		sendMu.Lock()
		defer sendMu.Unlock()
		return stream.Send(msg)
	}

	s.reg.add(agent.ID, safeSend)
	defer s.reg.remove(agent.ID)
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
		_ = s.svc.JobLog.IngestChunks(context.Background(), logBuf)
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
		}
		if u.ExitCode != 0 {
			ec := u.ExitCode
			in.ExitCode = &ec
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
