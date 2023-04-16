package handlers

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/verasthiago/verancial/worker/pkg/builder"
	"github.com/verasthiago/verancial/worker/pkg/report"
	"github.com/verasthiago/verancial/worker/pkg/types"
)

type CreateReportAPI interface {
	Handler(context context.Context, task *asynq.Task) error
}

type CreateReportHandler struct {
	builder.Builder
}

func (c *CreateReportHandler) InitFromBuilder(builder builder.Builder) *CreateReportHandler {
	c.Builder = builder
	return c
}

func (c *CreateReportHandler) Handler(context context.Context, task *asynq.Task) error {
	var payload types.QueuePayload
	if err := json.Unmarshal(task.Payload()[:], &payload); err != nil {
		return err
	}

	err, processor := report.GetReportProcessor(payload.BankName)
	if err != nil {
		return err
	}

	return processor.ProcessReport(payload.FilePath)
}
