package grpcapi

import (
	"context"
	"sync"
	"time"

	harmostv1 "github.com/harmost/proto/gen/harmost/v1"
	"github.com/harmost/hub/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	grpcstatus "google.golang.org/grpc/status"
)

const (
	logFlushInterval = 500 * time.Millisecond
	logBatchSize     = 500
	orgIDMetaKey     = "x-org-id"
)

// Connect is the bidirectional stream handler. Each connected agent gets one
// long-lived call to this method.
func (s *Server) Connect(stream grpc.BidiStreamingServer[harmostv1.AgentMessage, harmostv1.HubMessage]) error {
	ctx := stream.Context()

	orgID, err := orgIDFromCtx(ctx)
	if err != nil {
		return grpcstatus.Error(codes.Unauthenticated, "missing x-org-id metadata")
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

	agent, err := s.svc.Agent.Connect(ctx, orgID, domain.AgentConnectInput{
		Name:        hello.Hello.Name,
		Description: hello.Hello.Description,
		Version:     hello.Hello.Version,
		Hostname:    hello.Hello.Hostname,
	})
	if err != nil {
		return grpcstatus.Errorf(codes.Internal, "agent registration: %v", err)
	}

	// Serialise all hub→agent sends through a mutex so Dispatch (called from
	// other goroutines) and the recv loop can both write to the stream safely.
	var sendMu sync.Mutex
	safeSend := func(msg *harmostv1.HubMessage) error {
		sendMu.Lock()
		defer sendMu.Unlock()
		return stream.Send(msg)
	}

	s.reg.add(agent.ID, safeSend)
	defer s.reg.remove(agent.ID)
	defer s.svc.Agent.Disconnect(context.Background(), agent.ID)

	// Channels for the background recv goroutine.
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

	// Log chunk buffer — flushed every 500ms or when it hits 500 entries.
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
			s.handleMessage(ctx, agent.ID, r.msg, &logBuf, flush)

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
		_ = s.svc.Agent.HandleHeartbeat(ctx, agentID, p.Heartbeat.Timestamp.AsTime())

	case *harmostv1.AgentMessage_Pong:
		// latency tracking not implemented yet
	}
}

func orgIDFromCtx(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", grpcstatus.Error(codes.Unauthenticated, "no metadata")
	}
	vals := md.Get(orgIDMetaKey)
	if len(vals) == 0 || vals[0] == "" {
		return "", grpcstatus.Error(codes.Unauthenticated, "missing x-org-id")
	}
	return vals[0], nil
}
