package handlers

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/verasthiago/verancial/shared/models"
	"github.com/verasthiago/verancial/shared/types"
	"github.com/verasthiago/verancial/worker/pkg/builder"
	"github.com/verasthiago/verancial/worker/pkg/report"
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
	var payload types.QueuePayload
	var bankTransactions []interface{}
	var transactions []*models.Transaction

	if err := json.Unmarshal(task.Payload()[:], &payload); err != nil {
		return err
	}

	err, processor := report.GetReportProcessor(payload.BankName)
	if err != nil {
		return err
	}

	// TODO: Get current day and load all transactions from the previous day in user app
	err, bankTransactions = processor.LoadFromCSV(payload.FilePath)
	if err != nil {
		return err
	}

	err, transactions = processor.Process(bankTransactions)
	if err != nil {
		return err
	}

	return c.GetRepository().CreateTransactionInBatches(transactions)

}
