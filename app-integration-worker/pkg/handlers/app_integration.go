package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
	"github.com/verasthiago/verancial/app-integration-worker/pkg/builder"
	"github.com/verasthiago/verancial/app-integration-worker/pkg/generators"
	t "github.com/verasthiago/verancial/app-integration-worker/pkg/types"
	"github.com/verasthiago/verancial/shared/models"
	"github.com/verasthiago/verancial/shared/types"
)

type AppIntegrationAPI interface {
	Handler(context context.Context, task *asynq.Task) error
}

type AppIntegrationHandler struct {
	builder.Builder
}

func (c *AppIntegrationHandler) InitFromBuilder(builder builder.Builder) *AppIntegrationHandler {
	c.Builder = builder
	return c
}

func (c *AppIntegrationHandler) Handler(context context.Context, task *asynq.Task) error {
	var err error
	var user *models.User
	var appReport t.AppReport
	var lastTransaction time.Time
	var generator generators.AppReport
	var transactions []*models.Transaction
	var payload types.AppIntegrationQueuePayload

	if err = json.Unmarshal(task.Payload()[:], &payload); err != nil {
		return err
	}

	if user, err = c.GetRepository().GetUserByID(payload.UserId); err != nil {
		return err
	}

	if generator, err = generators.GetAppReportGenerator(payload.AppID); err != nil {
		return err
	}

	if lastTransaction, err = generator.GetLastTransaction(user.FinancialAppCredentials[payload.AppID], payload.BankId); err != nil {
		return err
	}

	if transactions, err = c.GetRepository().GetAllTransactionsFromUserBankAfterDate(user.ID, payload.BankId, lastTransaction); err != nil {
		return err
	}

	if appReport, err = generator.Generate(transactions); err != nil {
		return err
	}

	return generator.Submit(user, appReport)
}
