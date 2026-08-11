package events

import (
	"sync"
	"time"

	api "github.com/harmost/api/gen"
)

type EventType string

const (
	AgentConnected    EventType = "agent.connected"
	AgentDisconnected EventType = "agent.disconnected"
	AgentHeartbeat    EventType = "agent.heartbeat"
	AgentContainers   EventType = "agent.containers"
	JobStatus         EventType = "job.status"
	JobLog            EventType = "job.log"
)

// Event is the envelope written to each /ws frame. It stays hand-written rather
// than generated: OrgID is the bus routing key and never goes on the wire, and
// the generated HubEvent is a oneOf union that would be clumsy to publish into.
// The payloads below it are generated, so the front's types come from the same
// source. See libs/harmost-api/openapi.yaml.
type Event struct {
	Type    EventType `json:"type"`
	OrgID   string    `json:"-"`
	AgentID string    `json:"agent_id"`
	JobID   string    `json:"job_id,omitempty"`
	At      time.Time `json:"at"`
	Payload any       `json:"payload,omitempty"`
}

// Payload types for job.status and job.log events, generated from the OpenAPI
// contract shared with the front. Re-exported so publishers depend on the bus's
// vocabulary rather than importing the contract package directly.
type (
	JobStatusPayload  = api.JobStatusPayload
	JobLogPayload     = api.JobLogPayload
	LogLine           = api.LogLine
	JobState          = api.JobState
	LogStream         = api.LogStream
	ContainersPayload = api.ContainersPayload
	ContainerInfo     = api.ContainerInfo
)

type Bus struct {
	mu   sync.RWMutex
	subs map[string]map[int]chan Event
	next int
}

func New() *Bus {
	return &Bus{subs: make(map[string]map[int]chan Event)}
}

func (b *Bus) Publish(e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subs[e.OrgID] {
		select {
		case ch <- e:
		default:
		}
	}
}

func (b *Bus) Subscribe(orgID string) (<-chan Event, func()) {
	ch := make(chan Event, 64)
	b.mu.Lock()
	id := b.next
	b.next++
	if b.subs[orgID] == nil {
		b.subs[orgID] = make(map[int]chan Event)
	}
	b.subs[orgID][id] = ch
	b.mu.Unlock()

	return ch, func() {
		b.mu.Lock()
		delete(b.subs[orgID], id)
		b.mu.Unlock()
		close(ch)
	}
}
