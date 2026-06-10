package daemon

import (
	"context"
	"log"
	"time"

	"github.com/kardianos/service"
)

// AgentProgram implements service.Interface
type AgentProgram struct {
	cancel context.CancelFunc
}

func NewAgentProgram() *AgentProgram {
	return &AgentProgram{}
}

// Start is called by kardianos/service. It must be not block.
func (p *AgentProgram) Start(s service.Service) error {
	var ctx context.Context
	ctx, p.cancel = context.WithCancel(context.Background())

	// Start the actual work in a background goroutine
	go p.run(ctx)

	return nil
}

// TODO: Implement the gRPC reconnect loop:
//  1. Load credentials (hub address + OAuth token) from the OS config file written by `pair`.
//  2. Dial the hub over gRPC using those credentials.
//  3. Open the bidirectional stream.
//  4. Receive loop: block on stream.Recv(), dispatch incoming tasks/config to handlers.
//  5. Send: push task results/events back up via stream.Send().
//  6. On any error (stream closed, dial failed): reconnect with exponential back-off.
//  7. Respect ctx.Done() — break the whole loop cleanly when the daemon is stopped.
func (p *AgentProgram) run(ctx context.Context) {
	log.Println("Agent daemon started.")

	// A simple ticker to simulate work
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// This triggers when Stop() calls p.cancel()
			log.Println("Agent daemon shutting down gracefully...")
			return
		case <-ticker.C:
			log.Println("Agent is running... (waiting for hub tasks)")
		}
	}
}

// Stop is called by the OS or when you hit Ctrl+C in dev
func (p *AgentProgram) Stop(s service.Service) error {
	log.Println("Agent daemon stopping...")
	if p.cancel != nil {
		p.cancel()
	}

	//Give it a brief moment to finish logging the shutdown
	time.Sleep(100 * time.Millisecond)
	return nil
}
