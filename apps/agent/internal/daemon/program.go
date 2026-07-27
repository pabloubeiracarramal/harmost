package daemon

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/harmost/agent/internal/config"
	"github.com/harmost/agent/internal/docker"
	agentgrpc "github.com/harmost/agent/internal/grpc"
	"github.com/kardianos/service"
)

// AgentProgram implements service.Interface
type AgentProgram struct {
	cancel context.CancelFunc
}

func NewAgentProgram() *AgentProgram {
	return &AgentProgram{}
}

func (p *AgentProgram) Start(s service.Service) error {
	var ctx context.Context
	ctx, p.cancel = context.WithCancel(context.Background())
	go p.run(ctx)
	return nil
}

func (p *AgentProgram) run(ctx context.Context) {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("agent: no config found — run 'harmost pair <hub-url>' first: %v", err)
		return
	}

	target := cfg.GRPCAddr
	if target == "" {
		target = agentgrpc.Target(cfg.HubAddr)
	}
	log.Printf("agent: connecting to hub at %s", target)

	var mgr *docker.Manager
	if dock, err := docker.New(); err != nil {
		log.Printf("agent: docker unavailable, job execution disabled: %v", err)
	} else {
		if err := dock.Ping(ctx); err != nil {
			log.Printf("agent: docker daemon not reachable (jobs will fail until it is): %v", err)
		}
		mgr = docker.NewManager(dock)
	}

	client := agentgrpc.New(mgr)
	backoff := time.Second
	const maxBackoff = 60 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := client.Connect(ctx, target, cfg.Token, cfg.Insecure); err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("agent: connection lost (%v), reconnecting in %s", err, backoff)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
			backoff = min(time.Duration(float64(backoff)*math.Phi), maxBackoff)
			continue
		}
		backoff = time.Second
	}
}

func (p *AgentProgram) Stop(s service.Service) error {
	log.Println("agent: stopping...")
	if p.cancel != nil {
		p.cancel()
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}
