package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/verasthiago/verancial/app-integration-worker/pkg/builder"
	"github.com/verasthiago/verancial/app-integration-worker/pkg/generators"
	t "github.com/verasthiago/verancial/app-integration-worker/pkg/types"
	"github.com/verasthiago/verancial/shared/models"
	"github.com/verasthiago/verancial/shared/types"
)

type AppIntegrationAPI interface {
	HandlerAsync(context context.Context, task *asynq.Task) error
	HandlerSync(context *gin.Context) error
}

type AppIntegrationHandler struct {
	builder.Builder
}

func (a *AppIntegrationHandler) InitFromBuilder(builder builder.Builder) *AppIntegrationHandler {
	a.Builder = builder
	return a
}

func (a *AppIntegrationHandler) HandlerSync(context *gin.Context) error {
	var payload types.AppIntegrationQueuePayload

	fmt.Printf("\n[0/7] Parsing request payload...")
	if err := context.ShouldBindJSON(&payload); err != nil {
		return err
	}

	return a.Execute(payload)
}

func (a *AppIntegrationHandler) HandlerAsync(context context.Context, task *asynq.Task) error {
	var payload types.AppIntegrationQueuePayload

	fmt.Printf("[0/7] Parsing request payload...\n")
	if err := json.Unmarshal(task.Payload()[:], &payload); err != nil {
		return err
	}

	fmt.Printf("\npayload %+v\n", payload)

	return a.Execute(payload)
}

func (a *AppIntegrationHandler) Execute(payload types.AppIntegrationQueuePayload) error {
	var err error
	var user *models.User
	var appReport t.AppReport
	var lastTransaction time.Time
	var generator generators.AppReport
	var transactions []*models.Transaction

	defer func(startTime time.Time) {
		fmt.Printf("Total time %#v seconds\n", time.Since(startTime).Seconds())
	}(time.Now())

	fmt.Printf("[1/7] Getting user from database...\n")
	if user, err = a.GetRepository().GetUserByID(payload.UserId); err != nil {
		fmt.Printf("\nerr line 51 %+v\n", err)
		return err
	}

	fmt.Printf("[2/7] Geting AppReportGenerator...\n")
	if generator, err = generators.GetAppReportGenerator(payload.AppID); err != nil {
		fmt.Printf("\nerr line 56 %+v\n", err)
		return err
	}

	fmt.Printf("[3/7] Getting user app last transaction...\n")
	if lastTransaction, err = generator.GetLastTransaction(user.FinancialAppCredentials[payload.AppID], payload.BankId, payload.LastTransactionDate); err != nil {
		fmt.Printf("\nerr line 61 %+v\n", err)
		return err
	}
	fmt.Printf("\nlastTransaction %+v\n", lastTransaction)

	fmt.Printf("[4/7] Getting all transactions after last from app...\n")
	if transactions, err = a.GetRepository().GetAllTransactionsFromUserBankAfterDate(user.ID, payload.BankId, lastTransaction); err != nil {
		fmt.Printf("\nerr line 66 %+v\n", err)
		return err
	}

	fmt.Printf("[5/7] Generating App report...\n")
	if appReport, err = generator.Generate(transactions); err != nil {
		fmt.Printf("\nerr line 71 %+v\n", err)
		return err
	}

	fmt.Printf("[6/7] Submiting App report...\n")
	if err = generator.Submit(user, appReport); err != nil {
		fmt.Printf("\nerr line 79 %+v\n", err)
		return err
	}

	fmt.Printf("[7/7] DONE!\n")
	return nil
}
