package task

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	shared "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/shared/types"
)

type AsyncQueue struct {
	t *asynq.Client
}

func (a *AsyncQueue) InitFromFlags(sharedFlags *shared.SharedFlags) Task {
	redisConnection := asynq.RedisClientOpt{
		Addr: fmt.Sprintf("%+v:%+v", sharedFlags.QueueHost, sharedFlags.QueuePort),
	}

	t := asynq.NewClient(redisConnection)
	return &AsyncQueue{
		t,
	}
}

func (a *AsyncQueue) CreateReportAsync(payload types.ReportProcessQueuePayload) error {
	p, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(types.PatternReportProcess, p)
	//TODO: Define the right pattern of queue
	_, err = a.t.Enqueue(task, asynq.Queue("low"))
	return err
}

func (a *AsyncQueue) UpdateAppAsync(payload types.AppIntegrationQueuePayload) error {
	p, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(types.PatternAppIntegration, p)
	_, err = a.t.Enqueue(task, asynq.Queue("low"))
	return err
}
