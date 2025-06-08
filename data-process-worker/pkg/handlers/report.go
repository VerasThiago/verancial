package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/verasthiago/verancial/data-process-worker/pkg/builder"
	"github.com/verasthiago/verancial/data-process-worker/pkg/report"
	"github.com/verasthiago/verancial/shared/models"
	"github.com/verasthiago/verancial/shared/types"
)

type CreateReportAPI interface {
	HandlerSync(context *gin.Context) error
	HandlerAsync(context context.Context, task *asynq.Task) error
}
type CreateReportHandler struct {
	builder.Builder
}

func (c *CreateReportHandler) InitFromBuilder(builder builder.Builder) *CreateReportHandler {
	c.Builder = builder
	return c
}

func (c *CreateReportHandler) HandlerAsync(context context.Context, task *asynq.Task) error {
	var payload types.ReportProcessQueuePayload

	fmt.Printf("\n[0/7] Parsing request payload...")
	if err := json.Unmarshal(task.Payload()[:], &payload); err != nil {
		return err
	}

	return c.Execute(payload)
}

func (c *CreateReportHandler) HandlerSync(context *gin.Context) error {
	var payload types.ReportProcessQueuePayload

	fmt.Printf("\n[0/7] Parsing request payload...")
	if err := context.ShouldBindJSON(&payload); err != nil {
		return err
	}

	return c.Execute(payload)
}

func (c *CreateReportHandler) Execute(payload types.ReportProcessQueuePayload) error {
	var err error
	var bankTransactions []interface{}
	var transactions []*models.Transaction
	var lastDbTransaction *models.Transaction

	fmt.Printf("\n[1/7] Getting ReportProcessor for %s ...", payload.FilePath)
	processor, err := report.GetReportProcessor(payload.BankId)
	if err != nil {
		fmt.Printf("Error line 41: %+v\n", err)
		return err
	}

	fmt.Printf("\n[2/7] Loading transactions from the file...")
	// TODO: Get current day and load all transactions from the previous day in user app
	bankTransactions, err = processor.LoadFromCSV(payload.FilePath)
	if err != nil {
		fmt.Printf("Error line 50: %+v\n", err)
		return err
	}

	fmt.Printf("\n[3/7] Loading last database transaction...\n")
	if lastDbTransaction, err = c.GetRepository().GetLastTransactionFromUserBank(payload.UserId, payload.BankId); err != nil {
		fmt.Printf("Error line 55: %+v\n", err)
		return err
	}

	//TODO: Fix this workaround
	if lastDbTransaction == nil {
		fmt.Printf("\nlastDbTransaction is NULL\n")
		lastDbTransaction = &models.Transaction{Date: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)}
	}

	fmt.Printf("\n[4/7] Processing csv transations...\n")
	transactions, err = processor.Process(bankTransactions, &payload, lastDbTransaction)
	if err != nil {
		fmt.Printf("Error line 61: %+v\n", err)
		return err
	}

	fmt.Printf("\n[5/7] Setting transaction fingerprints for universal duplicate detection...\n")
	for _, tx := range transactions {
		tx.SetFingerprint()
	}

	fmt.Printf("\n[6/7] Saving new transactions to database\n")
	if err := c.GetRepository().CreateUniqueTransactionInBatches(transactions); err != nil {
		fmt.Printf("Error line 75: %+v\n", err)
		return err
	}

	fmt.Printf("\n[7/7] DONE!!\n")
	return err
}
