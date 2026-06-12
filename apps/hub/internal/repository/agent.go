package repository

import (
	"context"
	"time"

	"github.com/harmost/hub/internal/domain"
	"gorm.io/gorm"
)

type AgentRepo struct {
	db *gorm.DB
}

func (r *AgentRepo) Create(ctx context.Context, agent *domain.Agent) error {
	return r.db.WithContext(ctx).Create(agent).Error
}

func (r *AgentRepo) GetByID(ctx context.Context, orgID, id string) (*domain.Agent, error) {
	var a domain.Agent
	err := r.db.WithContext(ctx).
		First(&a, "id = ? AND org_id = ?", id, orgID).Error
	return &a, notFound(err)
}

func (r *AgentRepo) ListByOrg(ctx context.Context, orgID string) ([]domain.Agent, error) {
	var agents []domain.Agent
	err := r.db.WithContext(ctx).
		Where("org_id = ?", orgID).
		Order("created_at DESC").
		Find(&agents).Error
	return agents, err
}

func (r *AgentRepo) SetOnline(ctx context.Context, id string, at time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.Agent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":       domain.AgentStatusOnline,
			"last_seen_at": at,
		}).Error
}

func (r *AgentRepo) SetOffline(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.Agent{}).
		Where("id = ?", id).
		Update("status", domain.AgentStatusOffline).Error
}

// UpdateLastSeen refreshes last_seen_at on heartbeat without changing status.
func (r *AgentRepo) UpdateLastSeen(ctx context.Context, id string, at time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.Agent{}).
		Where("id = ?", id).
		Update("last_seen_at", at).Error
}
