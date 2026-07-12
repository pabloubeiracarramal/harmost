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
)

type Event struct {
	Type    EventType `json:"type"`
	OrgID   string    `json:"-"`
	AgentID string    `json:"agent_id"`
	At      time.Time `json:"at"`
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
