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
	q := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id)
	if orgID != "" {
		q = q.Where("org_id = ?", orgID)
	}
	return &a, notFound(q.First(&a).Error)
}

func (r *AgentRepo) ListByOrg(ctx context.Context, orgID string) ([]domain.Agent, error) {
	var agents []domain.Agent
	err := r.db.WithContext(ctx).
		Where("org_id = ? AND deleted_at IS NULL", orgID).
		Order("created_at DESC").
		Find(&agents).Error
	return agents, err
}

// Delete soft-deletes the agent (org-scoped) so job history referencing it
// via the jobs.agent_id FK stays intact.
func (r *AgentRepo) Delete(ctx context.Context, orgID, id string) error {
	result := r.db.WithContext(ctx).Model(&domain.Agent{}).
		Where("id = ? AND org_id = ? AND deleted_at IS NULL", id, orgID).
		Update("deleted_at", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *AgentRepo) SetOnline(ctx context.Context, id string, at time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.Agent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":       domain.AgentStatusOnline,
			"last_seen_at": at,
		}).Error
}

// SetOffline marks the agent offline and stamps last_seen_at so the orphan
// sweeper's grace clock starts at disconnect.
func (r *AgentRepo) SetOffline(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.Agent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":       domain.AgentStatusOffline,
			"last_seen_at": time.Now(),
		}).Error
}

// MarkAllOffline flips every online agent to offline. Called once at hub
// startup: after a crash the per-stream Disconnect never ran, and the sweeper
// only looks at offline agents. last_seen_at is left untouched so the grace
// clock reflects real last contact.
func (r *AgentRepo) MarkAllOffline(ctx context.Context) error {
	return r.db.WithContext(ctx).Model(&domain.Agent{}).
		Where("status = ?", domain.AgentStatusOnline).
		Update("status", domain.AgentStatusOffline).Error
}

// UpdateLastSeen refreshes last_seen_at on heartbeat without changing status.
func (r *AgentRepo) UpdateLastSeen(ctx context.Context, id string, at time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.Agent{}).
		Where("id = ?", id).
		Update("last_seen_at", at).Error
}

// UpdateMetrics stores the latest system metrics snapshot from a heartbeat.
func (r *AgentRepo) UpdateMetrics(ctx context.Context, id string, m domain.AgentMetrics, at time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.Agent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"cpu_usage_percent":  m.CpuUsagePercent,
			"memory_used_bytes":  m.MemoryUsedBytes,
			"memory_total_bytes": m.MemoryTotalBytes,
			"disk_used_bytes":    m.DiskUsedBytes,
			"disk_total_bytes":   m.DiskTotalBytes,
			"running_containers": m.RunningContainers,
			"last_seen_at":       at,
		}).Error
}

// UpdateOnConnect sets agent metadata from AgentHello and marks the agent online.
func (r *AgentRepo) UpdateOnConnect(ctx context.Context, id string, in domain.AgentConnectInput, at time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.Agent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"name":         in.Name,
			"description":  in.Description,
			"version":      in.Version,
			"hostname":     in.Hostname,
			"status":       domain.AgentStatusOnline,
			"last_seen_at": at,
		}).Error
}
