package events

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBus_OrgScoping(t *testing.T) {
	b := New()
	ch, unsub := b.Subscribe("org-1")
	defer unsub()

	b.Publish(Event{Type: AgentConnected, OrgID: "org-1", AgentID: "a1", At: time.Now()})
	b.Publish(Event{Type: AgentConnected, OrgID: "org-2", AgentID: "a2", At: time.Now()})

	require.Len(t, ch, 1)
	e := <-ch
	assert.Equal(t, "a1", e.AgentID)
}

func TestBus_Unsubscribe(t *testing.T) {
	b := New()
	ch, unsub := b.Subscribe("org-1")
	unsub()

	_, open := <-ch
	assert.False(t, open, "channel must be closed after unsubscribe")

	// Publishing after unsubscribe must not panic or block.
	b.Publish(Event{Type: AgentConnected, OrgID: "org-1", At: time.Now()})
}

func TestBus_DropWhenFull(t *testing.T) {
	b := New()
	ch, unsub := b.Subscribe("org-1")
	defer unsub()

	// One more than the subscriber buffer; the overflow must be dropped
	// without blocking Publish.
	for range 65 {
		b.Publish(Event{Type: AgentHeartbeat, OrgID: "org-1", At: time.Now()})
	}

	assert.Equal(t, 64, len(ch))
}
