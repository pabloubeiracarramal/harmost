package events

import (
	"sync"
	"time"
)

type EventType string

const (
	AgentConnected    EventType = "agent.connected"
	AgentDisconnected EventType = "agent.disconnected"
	AgentHeartbeat    EventType = "agent.heartbeat"
	JobStatus         EventType = "job.status"
	JobLog            EventType = "job.log"
)

type Event struct {
	Type    EventType `json:"type"`
	OrgID   string    `json:"-"`
	AgentID string    `json:"agent_id"`
	JobID   string    `json:"job_id,omitempty"`
	At      time.Time `json:"at"`
	Payload any       `json:"payload,omitempty"`
}

// JobStatusPayload is the payload of a job.status event.
type JobStatusPayload struct {
	State    string `json:"state"`
	Message  string `json:"message,omitempty"`
	ExitCode *int32 `json:"exit_code,omitempty"`
}

// LogLine is one log line inside a job.log event.
type LogLine struct {
	Line      string    `json:"line"`
	Stream    string    `json:"stream"`
	Sequence  int64     `json:"sequence"`
	Timestamp time.Time `json:"timestamp"`
}

// JobLogPayload is the payload of a job.log event — one event per job per
// flush batch, never per line.
type JobLogPayload struct {
	Lines []LogLine `json:"lines"`
}

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
