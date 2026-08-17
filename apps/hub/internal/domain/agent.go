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
	OrgID       string `gorm:"type:uuid;not null;index"`
	Name        string `gorm:"not null"`
	Description string
	Version     string
	Hostname    string
	Status      AgentStatus `gorm:"type:text;not null;default:'offline'"`
	LastSeenAt  *time.Time
	DeletedAt   *time.Time

	CpuUsagePercent   *float32
	MemoryUsedBytes   *int64
	MemoryTotalBytes  *int64
	DiskUsedBytes     *int64
	DiskTotalBytes    *int64
	RunningContainers *int32

	Org Org `gorm:"foreignKey:OrgID"`
}

type AgentConnectInput struct {
	Name        string
	Description string
	Version     string
	Hostname    string
}

type AgentMetrics struct {
	CpuUsagePercent   float32
	MemoryUsedBytes   int64
	MemoryTotalBytes  int64
	DiskUsedBytes     int64
	DiskTotalBytes    int64
	RunningContainers int32
}

type AgentRepository interface {
	Create(ctx context.Context, agent *Agent) error
	GetByID(ctx context.Context, orgID, id string) (*Agent, error)
	ListByOrg(ctx context.Context, orgID string) ([]Agent, error)
	SetOnline(ctx context.Context, id string, at time.Time) error
	SetOffline(ctx context.Context, id string) error
	MarkAllOffline(ctx context.Context) error
	UpdateLastSeen(ctx context.Context, id string, at time.Time) error
	UpdateOnConnect(ctx context.Context, id string, in AgentConnectInput, at time.Time) error
	UpdateMetrics(ctx context.Context, id string, m AgentMetrics, at time.Time) error
	Delete(ctx context.Context, orgID, id string) error
}

type AgentService interface {
	Connect(ctx context.Context, orgID string, in AgentConnectInput) (*Agent, error)
	UpdateOnConnect(ctx context.Context, id string, in AgentConnectInput) (*Agent, error)
	Disconnect(ctx context.Context, id string) error
	MarkAllOffline(ctx context.Context) error
	HandleHeartbeat(ctx context.Context, id string, m AgentMetrics, at time.Time) error
	List(ctx context.Context, orgID string) ([]Agent, error)
	GetByID(ctx context.Context, orgID, id string) (*Agent, error)
	Delete(ctx context.Context, orgID, id string) error
}
