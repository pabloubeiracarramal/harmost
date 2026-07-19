package grpc

import (
	"fmt"
	"testing"

	harmostv1 "github.com/harmost/proto/gen/harmost/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func statusMsg(jobID string, state harmostv1.JobState) *harmostv1.AgentMessage {
	return &harmostv1.AgentMessage{Payload: &harmostv1.AgentMessage_StatusUpdate{
		StatusUpdate: &harmostv1.JobStatusUpdate{JobId: jobID, State: state},
	}}
}

func logMsg(jobID string) *harmostv1.AgentMessage {
	return &harmostv1.AgentMessage{Payload: &harmostv1.AgentMessage_LogChunk{
		LogChunk: &harmostv1.LogChunk{JobId: jobID, Line: "line"},
	}}
}

func pongMsg() *harmostv1.AgentMessage {
	return &harmostv1.AgentMessage{Payload: &harmostv1.AgentMessage_Pong{
		Pong: &harmostv1.Pong{},
	}}
}

func TestSend_RoutesByPayloadType(t *testing.T) {
	c := New(nil)

	c.Send(statusMsg("j1", harmostv1.JobState_JOB_STATE_RUNNING))
	c.Send(logMsg("j1"))
	c.Send(pongMsg())

	assert.Equal(t, 2, len(c.statusCh), "status and pong route to statusCh")
	assert.Equal(t, 1, len(c.logCh))
}

func TestSend_DropsLogsWhenFull(t *testing.T) {
	c := New(nil)
	for range logBuffer {
		c.Send(logMsg("j1"))
	}

	c.Send(logMsg("j1")) // must not block
	assert.Equal(t, logBuffer, len(c.logCh))
}

func TestSend_DropsNonTerminalStatusWhenFull(t *testing.T) {
	c := New(nil)
	for i := 0; i < statusBuffer; i++ {
		c.Send(statusMsg(fmt.Sprintf("j%d", i), harmostv1.JobState_JOB_STATE_RUNNING))
	}

	c.Send(statusMsg("overflow", harmostv1.JobState_JOB_STATE_RUNNING)) // must not block
	require.Equal(t, statusBuffer, len(c.statusCh))

	first := <-c.statusCh
	assert.Equal(t, "j0", first.GetStatusUpdate().JobId, "oldest message kept — non-terminal overflow is dropped")
}

func TestSend_TerminalStatusEvictsOldestWhenFull(t *testing.T) {
	c := New(nil)
	for i := 0; i < statusBuffer; i++ {
		c.Send(statusMsg(fmt.Sprintf("j%d", i), harmostv1.JobState_JOB_STATE_RUNNING))
	}

	c.Send(statusMsg("terminal", harmostv1.JobState_JOB_STATE_SUCCEEDED))
	require.Equal(t, statusBuffer, len(c.statusCh))

	var last *harmostv1.AgentMessage
	for len(c.statusCh) > 0 {
		last = <-c.statusCh
	}
	assert.Equal(t, "terminal", last.GetStatusUpdate().JobId, "terminal status must be queued, oldest evicted")
}
