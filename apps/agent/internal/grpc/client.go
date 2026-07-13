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

type Client struct {
	mgr *docker.Manager // nil when Docker is unavailable on this host
}

func New(mgr *docker.Manager) *Client { return &Client{mgr: mgr} }

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

	hostname, _ := os.Hostname()
	if err := stream.Send(&harmostv1.AgentMessage{
		Payload: &harmostv1.AgentMessage_Hello{
			Hello: &harmostv1.AgentHello{
				Name:        hostname,
				Description: fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
				Version:     "0.1.0",
				Hostname:    hostname,
			},
		},
	}); err != nil {
		return fmt.Errorf("send hello: %w", err)
	}

	log.Printf("agent: connected to hub")

	// gRPC streams allow only one sending goroutine. Jobs run concurrently
	// and report status/logs at any time, so everything funnels through
	// sendCh and the select loop below is the sole caller of stream.Send.
	sendCh := make(chan *harmostv1.AgentMessage, 256)
	send := func(msg *harmostv1.AgentMessage) {
		select {
		case sendCh <- msg:
		default:
			log.Printf("agent: send buffer full, dropping message")
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
			c.handleMessage(ctx, msg, send)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-recvCh:
			return err
		case msg := <-sendCh:
			if err := stream.Send(msg); err != nil {
				return fmt.Errorf("send: %w", err)
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

func (c *Client) handleMessage(ctx context.Context, msg *harmostv1.HubMessage, send docker.SendFunc) {
	switch payload := msg.Payload.(type) {
	case *harmostv1.HubMessage_DispatchJob:
		log.Printf("agent: received job %s", payload.DispatchJob.JobId)
		if c.mgr == nil {
			send(&harmostv1.AgentMessage{Payload: &harmostv1.AgentMessage_StatusUpdate{
				StatusUpdate: &harmostv1.JobStatusUpdate{
					JobId:     payload.DispatchJob.JobId,
					State:     harmostv1.JobState_JOB_STATE_FAILED,
					Timestamp: timestamppb.Now(),
					Message:   "docker is not available on this agent",
				},
			}})
			return
		}
		c.mgr.Dispatch(ctx, payload.DispatchJob.JobId, payload.DispatchJob.Spec, send)
	case *harmostv1.HubMessage_CancelJob:
		if c.mgr == nil || !c.mgr.Cancel(payload.CancelJob.JobId) {
			log.Printf("agent: cancel for unknown job %s", payload.CancelJob.JobId)
		}
	case *harmostv1.HubMessage_Ping:
		send(&harmostv1.AgentMessage{Payload: &harmostv1.AgentMessage_Pong{
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
