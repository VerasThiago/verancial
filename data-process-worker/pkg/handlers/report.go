package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
	var bankTransactions []interface{}
	var transactions []*models.Transaction
	var lastDbTransaction *models.Transaction
	var payload types.ReportProcessQueuePayload

	fmt.Printf("\n[0/6] Parsing request payload...")
	if err := json.Unmarshal(task.Payload()[:], &payload); err != nil {
		return err
	}

	fmt.Printf("\n[1/6] Getting ReportProcessor for %s ...", payload.FilePath)
	processor, err := report.GetReportProcessor(payload.BankId)
	if err != nil {
		fmt.Printf("Error line 41: %+v\n", err)
		return err
	}

	fmt.Printf("\n[2/6] Loading transactions from the file...")
	// TODO: Get current day and load all transactions from the previous day in user app
	bankTransactions, err = processor.LoadFromCSV(payload.FilePath)
	fmt.Printf("\nbankTransactions[0] %+v\n", bankTransactions[0])
	if err != nil {
		fmt.Printf("Error line 50: %+v\n", err)
		return err
	}

	fmt.Printf("\n[3/6] Loading last database transaction...\n")
	if lastDbTransaction, err = c.GetRepository().GetLastTransactionFromUserBank(payload.UserId, payload.BankId); err != nil {
		fmt.Printf("Error line 55: %+v\n", err)
		return err
	}

	//TODO: Fix this workaround
	if lastDbTransaction == nil {
		fmt.Printf("\nlastDbTransaction is NULL\n")
		lastDbTransaction = &models.Transaction{Date: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)}
	}

	fmt.Printf("\nlastDbTransaction %+v\n", lastDbTransaction.Date)

	fmt.Printf("\n[4/6] Processing csv transations...\n")
	transactions, err = processor.Process(bankTransactions, &payload, lastDbTransaction)
	if err != nil {
		fmt.Printf("Error line 61: %+v\n", err)
		return err
	}

	fmt.Printf("\n[5/6] Saving new transactions to database...\n")
	err = c.GetRepository().CreateTransactionInBatches(transactions)
	if err != nil {
		fmt.Printf("Error line 75: %+v\n", err)
	}

	fmt.Printf("\n[6/6] DONE!!\n")
	return err
}
