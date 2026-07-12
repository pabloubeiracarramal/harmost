package grpc

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/harmost/agent/internal/metrics"
	harmostv1 "github.com/harmost/proto/gen/harmost/v1"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Client struct{}

func New() *Client { return &Client{} }

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
			handleMessage(msg)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-recvCh:
			return err
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

func handleMessage(msg *harmostv1.HubMessage) {
	switch payload := msg.Payload.(type) {
	case *harmostv1.HubMessage_DispatchJob:
		log.Printf("agent: received job %s", payload.DispatchJob.JobId)
	case *harmostv1.HubMessage_Ping:
		log.Printf("agent: ping from hub")
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
