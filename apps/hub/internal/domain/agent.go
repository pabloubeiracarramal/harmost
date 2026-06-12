package domain

import (
	"context"
	"time"
)

type AgentStatus string

const (
	AgentStatusOnline  AgentStatus = "online"
	AgentStatusOffline AgentStatus = "offline"
)

type Agent struct {
	Model
	OrgID       string      `gorm:"type:uuid;not null;index"`
	Name        string      `gorm:"not null"`
	Description string
	Version     string
	Hostname    string
	Status      AgentStatus `gorm:"type:text;not null;default:'offline'"`
	LastSeenAt  *time.Time

	Org Org `gorm:"foreignKey:OrgID"`
}

type AgentConnectInput struct {
	Name        string
	Description string
	Version     string
	Hostname    string
}

type AgentRepository interface {
	Create(ctx context.Context, agent *Agent) error
	GetByID(ctx context.Context, orgID, id string) (*Agent, error)
	ListByOrg(ctx context.Context, orgID string) ([]Agent, error)
	SetOnline(ctx context.Context, id string, at time.Time) error
	SetOffline(ctx context.Context, id string) error
	UpdateLastSeen(ctx context.Context, id string, at time.Time) error
}

type AgentService interface {
	Connect(ctx context.Context, orgID string, in AgentConnectInput) (*Agent, error)
	Disconnect(ctx context.Context, id string) error
	HandleHeartbeat(ctx context.Context, id string, at time.Time) error
	List(ctx context.Context, orgID string) ([]Agent, error)
	GetByID(ctx context.Context, orgID, id string) (*Agent, error)
}
