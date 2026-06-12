package grpcapi

import (
	"context"
	"fmt"
	"sync"

	harmostv1 "github.com/harmost/proto/gen/harmost/v1"
	"github.com/harmost/hub/internal/domain"
	"github.com/harmost/hub/internal/service"
	"google.golang.org/grpc"
)

// Server implements AgentServiceServer and holds the active stream registry.
type Server struct {
	harmostv1.UnimplementedAgentServiceServer
	svc *service.Services
	reg *registry
}

func New(svc *service.Services) *Server {
	return &Server{svc: svc, reg: newRegistry()}
}

// Register wires this server into a gRPC server instance.
func (s *Server) Register(g *grpc.Server) {
	harmostv1.RegisterAgentServiceServer(g, s)
}

// Dispatch sends a job to a currently-connected agent.
// Returns an error if the agent has no active stream.
func (s *Server) Dispatch(ctx context.Context, agentID string, job *domain.Job) error {
	send, ok := s.reg.get(agentID)
	if !ok {
		return fmt.Errorf("agent %s is not connected", agentID)
	}
	return send(&harmostv1.HubMessage{
		Payload: &harmostv1.HubMessage_DispatchJob{
			DispatchJob: jobToProto(job),
		},
	})
}

// ─── registry ────────────────────────────────────────────────────────────────

type sendFn func(*harmostv1.HubMessage) error

type registry struct {
	mu      sync.RWMutex
	streams map[string]sendFn
}

func newRegistry() *registry {
	return &registry{streams: make(map[string]sendFn)}
}

func (r *registry) add(agentID string, fn sendFn) {
	r.mu.Lock()
	r.streams[agentID] = fn
	r.mu.Unlock()
}

func (r *registry) remove(agentID string) {
	r.mu.Lock()
	delete(r.streams, agentID)
	r.mu.Unlock()
}

func (r *registry) get(agentID string) (sendFn, bool) {
	r.mu.RLock()
	fn, ok := r.streams[agentID]
	r.mu.RUnlock()
	return fn, ok
}
