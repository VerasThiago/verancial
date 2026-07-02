package task

import (
	"encoding/json"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	shared "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/shared/types"
)

// newTestQueue spins up an in-memory Redis (miniredis) and wires an
// AsyncQueue at it, along with an asynq.Inspector for asserting on what
// actually got enqueued.
func newTestQueue(t *testing.T) (Task, *asynq.Inspector) {
	t.Helper()

	mr := miniredis.RunT(t)

	sharedFlags := &shared.SharedFlags{
		QueueHost: mr.Host(),
		QueuePort: mr.Port(),
	}

	q := new(AsyncQueue).InitFromFlags(sharedFlags)

	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: mr.Addr()})
	t.Cleanup(func() { inspector.Close() })

	return q, inspector
}

func TestAsyncQueue_CreateReportAsync(t *testing.T) {
	q, inspector := newTestQueue(t)

	payload := types.ReportProcessQueuePayload{
		UserId:   "user-1",
		BankId:   "bank-1",
		FilePath: "/tmp/upload_user-1_statement.csv",
	}

	require.NoError(t, q.CreateReportAsync(payload))

	tasks, err := inspector.ListPendingTasks("low")
	require.NoError(t, err)
	require.Len(t, tasks, 1)

	assert.Equal(t, types.PatternReportProcess, tasks[0].Type)

	var got types.ReportProcessQueuePayload
	require.NoError(t, json.Unmarshal(tasks[0].Payload, &got))
	assert.Equal(t, payload, got)
}

func TestAsyncQueue_UpdateAppAsync(t *testing.T) {
	q, inspector := newTestQueue(t)

	payload := types.AppIntegrationQueuePayload{
		UserId:              "user-1",
		AppID:                "9f3df639-5d19-4d4d-baeb-9e9742107617",
		BankId:               "bank-1",
		LastTransactionDate:  "January 2 2006",
	}

	require.NoError(t, q.UpdateAppAsync(payload))

	tasks, err := inspector.ListPendingTasks("low")
	require.NoError(t, err)
	require.Len(t, tasks, 1)

	assert.Equal(t, types.PatternAppIntegration, tasks[0].Type)

	var got types.AppIntegrationQueuePayload
	require.NoError(t, json.Unmarshal(tasks[0].Payload, &got))
	assert.Equal(t, payload, got)
}

func TestAsyncQueue_EnqueuesBothTaskTypesIndependently(t *testing.T) {
	q, inspector := newTestQueue(t)

	require.NoError(t, q.CreateReportAsync(types.ReportProcessQueuePayload{UserId: "u1"}))
	require.NoError(t, q.UpdateAppAsync(types.AppIntegrationQueuePayload{UserId: "u2"}))

	tasks, err := inspector.ListPendingTasks("low")
	require.NoError(t, err)
	require.Len(t, tasks, 2)

	types_ := []string{tasks[0].Type, tasks[1].Type}
	assert.Contains(t, types_, types.PatternReportProcess)
	assert.Contains(t, types_, types.PatternAppIntegration)
}

