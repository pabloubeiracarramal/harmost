package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/harmost/hub/internal/domain"
	"gorm.io/gorm"
)

// ─── AgentTokenRepo ───────────────────────────────────────────────────────────

type AgentTokenRepo struct {
	db *gorm.DB
}

func (r *AgentTokenRepo) Create(ctx context.Context, t *domain.AgentToken) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *AgentTokenRepo) GetByHash(ctx context.Context, hash string) (*domain.AgentToken, error) {
	var t domain.AgentToken
	err := r.db.WithContext(ctx).
		First(&t, "token_hash = ? AND revoked_at IS NULL", hash).Error
	return &t, notFound(err)
}

func (r *AgentTokenRepo) TouchLastUsed(ctx context.Context, id string, at time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.AgentToken{}).
		Where("id = ?", id).
		Update("last_used_at", at).Error
}

func (r *AgentTokenRepo) ListByOrg(ctx context.Context, orgID string) ([]domain.AgentToken, error) {
	var tokens []domain.AgentToken
	err := r.db.WithContext(ctx).
		Where("org_id = ? AND revoked_at IS NULL", orgID).
		Order("created_at DESC").
		Find(&tokens).Error
	return tokens, err
}

func (r *AgentTokenRepo) Revoke(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.AgentToken{}).
		Where("id = ?", id).
		Update("revoked_at", now).Error
}

// ─── DeviceCodeRepo ───────────────────────────────────────────────────────────

type DeviceCodeRepo struct {
	db *gorm.DB
}

func (r *DeviceCodeRepo) Create(ctx context.Context, d *domain.DeviceCode) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *DeviceCodeRepo) GetByDeviceCode(ctx context.Context, deviceCode string) (*domain.DeviceCode, error) {
	var d domain.DeviceCode
	err := r.db.WithContext(ctx).First(&d, "device_code = ?", deviceCode).Error
	return &d, notFound(err)
}

func (r *DeviceCodeRepo) GetByUserCode(ctx context.Context, userCode string) (*domain.DeviceCode, error) {
	var d domain.DeviceCode
	err := r.db.WithContext(ctx).First(&d, "user_code = ?", userCode).Error
	return &d, notFound(err)
}

func (r *DeviceCodeRepo) Approve(ctx context.Context, deviceCode, orgID, token string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&domain.DeviceCode{}).
		Where("device_code = ? AND approved_at IS NULL AND expires_at > now()", deviceCode).
		Updates(map[string]any{
			"org_id":      orgID,
			"token":       token,
			"approved_at": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("device code not found or already approved")
	}
	return nil
}

// ConsumeToken retrieves the plaintext token once and clears it from the row.
func (r *DeviceCodeRepo) ConsumeToken(ctx context.Context, deviceCode string) (string, error) {
	var d domain.DeviceCode
	err := r.db.WithContext(ctx).First(&d, "device_code = ?", deviceCode).Error
	if err != nil {
		return "", notFound(err)
	}
	if d.Token == nil {
		return "", nil
	}
	tok := *d.Token
	r.db.WithContext(ctx).Model(&d).Update("token", nil)
	return tok, nil
}
