package handlers

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/verasthiago/verancial/data-process-worker/pkg/builder"
	"github.com/verasthiago/verancial/data-process-worker/pkg/report"
	"github.com/verasthiago/verancial/shared/models"
	"github.com/verasthiago/verancial/shared/types"
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
	var err error
	var payload types.ReportProcessQueuePayload
	var bankTransactions []interface{}
	var transactions []*models.Transaction

	if err := json.Unmarshal(task.Payload()[:], &payload); err != nil {
		return err
	}

	processor, err := report.GetReportProcessor(payload.BankId)
	if err != nil {
		return err
	}

	// TODO: Get current day and load all transactions from the previous day in user app
	bankTransactions, err = processor.LoadFromCSV(payload.FilePath)
	if err != nil {
		return err
	}

	transactions, err = processor.Process(bankTransactions, &payload)
	if err != nil {
		return err
	}

	return c.GetRepository().CreateTransactionInBatches(transactions)
}
