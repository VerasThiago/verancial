package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"os"
	"time"

	"github.com/hibiken/asynq"
	"github.com/verasthiago/verancial/app-integration-worker/pkg/builder"
	"github.com/verasthiago/verancial/app-integration-worker/pkg/helper"
	"github.com/verasthiago/verancial/app-integration-worker/pkg/helper/csvcreator"
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

func tmp(fileName string, appReport [][]string) error {
	// Create a new CSV file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a new CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the data rows
	for _, row := range appReport {
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func (c *AppIntegrationHandler) Handler(context context.Context, task *asynq.Task) error {
	var err error
	var user *models.User
	var appReport [][]string
	var lastTransaction time.Time
	var transactions []*models.Transaction
	var payload types.AppIntegrationQueuePayload

	if err = json.Unmarshal(task.Payload()[:], &payload); err != nil {
		return err
	}

	if user, err = c.GetRepository().GetUserByID(payload.UserId); err != nil {
		return err
	}

	if lastTransaction, err = helper.GetLastTransaction(user.FinancialAppCredentials[payload.AppID], payload.BankId); err != nil {
		return err
	}

	if transactions, err = c.GetRepository().GetAllTransactionsFromUserBankAfterDate(user.ID, payload.BankId, lastTransaction); err != nil {
		return err
	}

	//TODO: Use interface to get CSV
	if appReport, err = csvcreator.CreateCSVFromTransactions(transactions); err != nil {
		return err
	}

	//TODO: Remove this tmp func and inject on user app
	if err = tmp(user.Email+"-"+string(payload.AppID)+".csv", appReport); err != nil {
		return err
	}

	return nil
}
