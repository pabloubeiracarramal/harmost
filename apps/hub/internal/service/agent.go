package service

import (
	"context"
	"time"

	"github.com/harmost/hub/internal/domain"
	"github.com/harmost/hub/internal/repository"
	"gorm.io/gorm"
)

type AgentService struct {
	db        *gorm.DB
	agentRepo domain.AgentRepository
}

// Connect upserts an agent record from an AgentHello and marks it online.
// Agents are identified by (orgID, name) — reconnecting with the same name
// updates the existing record rather than creating a new one.
func (s *AgentService) Connect(ctx context.Context, orgID string, in domain.AgentConnectInput) (*domain.Agent, error) {
	var agent *domain.Agent
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepos := repository.New(tx)

		agents, err := txRepos.Agent.ListByOrg(ctx, orgID)
		if err != nil {
			return err
		}

		var existing *domain.Agent
		for i := range agents {
			if agents[i].Name == in.Name {
				existing = &agents[i]
				break
			}
		}

		now := time.Now()
		if existing == nil {
			a := &domain.Agent{
				OrgID:       orgID,
				Name:        in.Name,
				Description: in.Description,
				Version:     in.Version,
				Hostname:    in.Hostname,
				Status:      domain.AgentStatusOnline,
				LastSeenAt:  &now,
			}
			if err := txRepos.Agent.Create(ctx, a); err != nil {
				return err
			}
			agent = a
		} else {
			existing.Description = in.Description
			existing.Version = in.Version
			existing.Hostname = in.Hostname
			if err := txRepos.Agent.SetOnline(ctx, existing.ID, now); err != nil {
				return err
			}
			agent = existing
		}
		return nil
	})
	return agent, err
}

// UpdateOnConnect updates an existing agent record (created at pairing) with live
// metadata from AgentHello and marks it online.
func (s *AgentService) UpdateOnConnect(ctx context.Context, id string, in domain.AgentConnectInput) (*domain.Agent, error) {
	now := time.Now()
	if err := s.agentRepo.UpdateOnConnect(ctx, id, in, now); err != nil {
		return nil, err
	}
	return s.agentRepo.GetByID(ctx, "", id)
}

// Disconnect marks the agent offline when its gRPC stream closes.
func (s *AgentService) Disconnect(ctx context.Context, id string) error {
	return s.agentRepo.SetOffline(ctx, id)
}

// MarkAllOffline resets every agent to offline at hub startup, when no
// streams exist yet — recovers agents left online by a crash.
func (s *AgentService) MarkAllOffline(ctx context.Context) error {
	return s.agentRepo.MarkAllOffline(ctx)
}

// HandleHeartbeat stores the latest metrics snapshot and refreshes last_seen_at.
func (s *AgentService) HandleHeartbeat(ctx context.Context, id string, m domain.AgentMetrics, at time.Time) error {
	return s.agentRepo.UpdateMetrics(ctx, id, m, at)
}

// List returns all agents belonging to an org.
func (s *AgentService) List(ctx context.Context, orgID string) ([]domain.Agent, error) {
	return s.agentRepo.ListByOrg(ctx, orgID)
}

// GetByID returns a single agent scoped to an org.
func (s *AgentService) GetByID(ctx context.Context, orgID, id string) (*domain.Agent, error) {
	return s.agentRepo.GetByID(ctx, orgID, id)
}
