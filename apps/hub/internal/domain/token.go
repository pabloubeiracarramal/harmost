package domain

import (
	"context"
	"time"
)

type AgentToken struct {
	Model
	OrgID       string     `gorm:"type:uuid;not null;index"`
	AgentID     *string    `gorm:"type:uuid"`
	Name        string     `gorm:"not null"`
	TokenHash   string     `gorm:"not null;uniqueIndex"`
	CreatedByID string     `gorm:"type:uuid;not null"`
	LastUsedAt  *time.Time
	RevokedAt   *time.Time
}

type DeviceCode struct {
	DeviceCode string     `gorm:"primaryKey"`
	UserCode   string     `gorm:"uniqueIndex;not null"`
	OrgID      *string    `gorm:"type:uuid"`
	Token      *string
	ExpiresAt  time.Time  `gorm:"not null"`
	ApprovedAt *time.Time
	CreatedAt  time.Time
}

type AgentTokenRepository interface {
	Create(ctx context.Context, t *AgentToken) error
	GetByHash(ctx context.Context, hash string) (*AgentToken, error)
	TouchLastUsed(ctx context.Context, id string, at time.Time) error
	ListByOrg(ctx context.Context, orgID string) ([]AgentToken, error)
	Revoke(ctx context.Context, orgID, id string) error
}

type DeviceCodeRepository interface {
	Create(ctx context.Context, d *DeviceCode) error
	GetByDeviceCode(ctx context.Context, deviceCode string) (*DeviceCode, error)
	GetByUserCode(ctx context.Context, userCode string) (*DeviceCode, error)
	Approve(ctx context.Context, deviceCode string, orgID string, token string) error
	ConsumeToken(ctx context.Context, deviceCode string) (string, error)
}

type AgentTokenService interface {
	Generate(ctx context.Context, orgID, name, createdByID string) (plaintext string, err error)
	Validate(ctx context.Context, token string) (orgID, agentID string, err error)
	List(ctx context.Context, orgID string) ([]AgentToken, error)
	Revoke(ctx context.Context, orgID, id string) error
}

type DeviceFlowService interface {
	Initiate(ctx context.Context) (*DeviceCode, error)
	Approve(ctx context.Context, userCode string, orgID string, userID string) error
	Poll(ctx context.Context, deviceCode string) (token string, pending bool, err error)
}
