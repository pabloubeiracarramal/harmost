package events

import (
	"encoding/json"
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

// Guards the /ws frame format after the payload types became generated from
// libs/harmost-api/openapi.yaml (ADR 0010). The front's HubEvent union is
// generated from the same document, so these shapes must not drift.
func TestEventWireFormat(t *testing.T) {
	at := time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)

	t.Run("agent event carries no payload and never leaks OrgID", func(t *testing.T) {
		b, err := json.Marshal(Event{
			Type: AgentHeartbeat, OrgID: "org-secret", AgentID: "a1", At: at,
		})
		require.NoError(t, err)
		assert.JSONEq(t, `{"type":"agent.heartbeat","agent_id":"a1",
			"at":"2026-08-08T12:00:00Z"}`, string(b))
		assert.NotContains(t, string(b), "org-secret")
	})

	t.Run("job.status omits an empty message but keeps a zero exit code", func(t *testing.T) {
		exit := int32(0)
		b, err := json.Marshal(Event{
			Type: JobStatus, OrgID: "o", AgentID: "a1", JobID: "j1", At: at,
			Payload: JobStatusPayload{State: "succeeded", ExitCode: &exit},
		})
		require.NoError(t, err)
		assert.JSONEq(t, `{"type":"job.status","agent_id":"a1","job_id":"j1",
			"at":"2026-08-08T12:00:00Z",
			"payload":{"state":"succeeded","exit_code":0}}`, string(b))
	})

	t.Run("job.log batches lines", func(t *testing.T) {
		b, err := json.Marshal(Event{
			Type: JobLog, OrgID: "o", AgentID: "a1", JobID: "j1", At: at,
			Payload: JobLogPayload{Lines: []LogLine{
				{Line: "hi", Stream: "stdout", Sequence: 1, Timestamp: at},
			}},
		})
		require.NoError(t, err)
		assert.JSONEq(t, `{"type":"job.log","agent_id":"a1","job_id":"j1",
			"at":"2026-08-08T12:00:00Z","payload":{"lines":[
			{"line":"hi","stream":"stdout","sequence":1,
			 "timestamp":"2026-08-08T12:00:00Z"}]}}`, string(b))
	})
}
