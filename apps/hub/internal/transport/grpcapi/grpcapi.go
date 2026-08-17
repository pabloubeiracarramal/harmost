package grpcapi

import (
	"context"
	"fmt"
	"sync"

	"github.com/harmost/hub/internal/domain"
	"github.com/harmost/hub/internal/events"
	"github.com/harmost/hub/internal/service"
	harmostv1 "github.com/harmost/proto/gen/harmost/v1"
	"google.golang.org/grpc"
)

// Server implements AgentServiceServer and holds the active stream registry.
type Server struct {
	harmostv1.UnimplementedAgentServiceServer
	svc *service.Services
	reg *registry
	bus *events.Bus
}

func New(svc *service.Services, bus *events.Bus) *Server {
	return &Server{svc: svc, reg: newRegistry(), bus: bus}
}

// Register wires this server into a gRPC server instance.
func (s *Server) Register(g *grpc.Server) {
	harmostv1.RegisterAgentServiceServer(g, s)
}

// Connected reports whether an agent currently has an active stream.
func (s *Server) Connected(agentID string) bool {
	_, ok := s.reg.get(agentID)
	return ok
}

// Kick tells agentID's active stream, if any, that it has been unpaired —
// so it stops reconnecting on its own rather than just seeing a generic
// disconnect and retrying forever with a now-revoked token — then
// force-closes the stream so it doesn't linger until the agent notices on
// its own.
func (s *Server) Kick(agentID string) {
	if send, ok := s.reg.get(agentID); ok {
		_ = send(&harmostv1.HubMessage{
			Payload: &harmostv1.HubMessage_Unpair{Unpair: &harmostv1.Unpair{}},
		})
	}
	s.reg.kick(agentID)
}

// Dispatch sends a job to a currently-connected agent.
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

// Cancel asks a currently-connected agent to cancel a job.
func (s *Server) Cancel(ctx context.Context, agentID, jobID string) error {
	send, ok := s.reg.get(agentID)
	if !ok {
		return fmt.Errorf("agent %s is not connected", agentID)
	}
	return send(&harmostv1.HubMessage{
		Payload: &harmostv1.HubMessage_CancelJob{
			CancelJob: &harmostv1.CancelJobRequest{JobId: jobID},
		},
	})
}

// WatchContainers registers a front client's interest in agentID's running
// containers. Only the first watcher for a given agent actually asks it to
// start pushing — later ones ride the same stream of updates. An agent that
// isn't connected yet is not an error: Connect (connect.go) resends the
// watch on reconnect if the refcount is still non-zero.
func (s *Server) WatchContainers(ctx context.Context, agentID string) error {
	if first := s.reg.addWatcher(agentID); !first {
		return nil
	}
	send, ok := s.reg.get(agentID)
	if !ok {
		return nil
	}
	return send(&harmostv1.HubMessage{
		Payload: &harmostv1.HubMessage_WatchContainers{
			WatchContainers: &harmostv1.WatchContainersRequest{},
		},
	})
}

// UnwatchContainers releases one front client's interest in agentID's
// running containers. Only the last watcher going away actually asks the
// agent to stop pushing.
func (s *Server) UnwatchContainers(ctx context.Context, agentID string) error {
	if last := s.reg.removeWatcher(agentID); !last {
		return nil
	}
	send, ok := s.reg.get(agentID)
	if !ok {
		return nil
	}
	return send(&harmostv1.HubMessage{
		Payload: &harmostv1.HubMessage_UnwatchContainers{
			UnwatchContainers: &harmostv1.UnwatchContainersRequest{},
		},
	})
}

// ContainerAction asks a currently-connected agent to perform a lifecycle
// action on a container. Unlike Watch/Unwatch, an agent that isn't
// connected genuinely can't do anything — same as Dispatch/Cancel.
func (s *Server) ContainerAction(ctx context.Context, agentID, containerID, action string) error {
	send, ok := s.reg.get(agentID)
	if !ok {
		return fmt.Errorf("agent %s is not connected", agentID)
	}
	protoAction, ok := containerActionToProto[action]
	if !ok {
		return fmt.Errorf("unknown container action %q", action)
	}
	return send(&harmostv1.HubMessage{
		Payload: &harmostv1.HubMessage_ContainerAction{
			ContainerAction: &harmostv1.ContainerActionRequest{
				ContainerId: containerID,
				Action:      protoAction,
			},
		},
	})
}

var containerActionToProto = map[string]harmostv1.ContainerAction{
	"start":   harmostv1.ContainerAction_CONTAINER_ACTION_START,
	"stop":    harmostv1.ContainerAction_CONTAINER_ACTION_STOP,
	"restart": harmostv1.ContainerAction_CONTAINER_ACTION_RESTART,
	"remove":  harmostv1.ContainerAction_CONTAINER_ACTION_REMOVE,
}

// ─── registry ────────────────────────────────────────────────────────────────

type sendFn func(*harmostv1.HubMessage) error

type registry struct {
	mu       sync.RWMutex
	streams  map[string]sendFn
	kills    map[string]chan struct{}
	watchers map[string]int
}

func newRegistry() *registry {
	return &registry{
		streams:  make(map[string]sendFn),
		kills:    make(map[string]chan struct{}),
		watchers: make(map[string]int),
	}
}

// add registers the stream's send function and returns a channel that closes
// when kick is called for this agent.
func (r *registry) add(agentID string, fn sendFn) chan struct{} {
	r.mu.Lock()
	defer r.mu.Unlock()
	kill := make(chan struct{})
	r.streams[agentID] = fn
	r.kills[agentID] = kill
	return kill
}

func (r *registry) remove(agentID string) {
	r.mu.Lock()
	delete(r.streams, agentID)
	delete(r.kills, agentID)
	r.mu.Unlock()
}

func (r *registry) get(agentID string) (sendFn, bool) {
	r.mu.RLock()
	fn, ok := r.streams[agentID]
	r.mu.RUnlock()
	return fn, ok
}

// kick force-closes agentID's kill channel, if it has an active stream.
func (r *registry) kick(agentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	kill, ok := r.kills[agentID]
	if !ok {
		return
	}
	close(kill)
	delete(r.kills, agentID)
}

// addWatcher increments agentID's watcher count and reports whether this
// was the first watcher (0→1).
func (r *registry) addWatcher(agentID string) (first bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	first = r.watchers[agentID] == 0
	r.watchers[agentID]++
	return first
}

// removeWatcher decrements agentID's watcher count and reports whether this
// was the last watcher (1→0). A count already at zero is left alone.
func (r *registry) removeWatcher(agentID string) (last bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.watchers[agentID] <= 0 {
		return false
	}
	r.watchers[agentID]--
	if r.watchers[agentID] == 0 {
		delete(r.watchers, agentID)
		last = true
	}
	return last
}

// watcherCount reports how many front clients are currently watching
// agentID's containers.
func (r *registry) watcherCount(agentID string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.watchers[agentID]
}
