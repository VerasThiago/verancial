package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/verasthiago/verancial/api/pkg/builder"
	"github.com/verasthiago/verancial/shared/models"
)

type BankStatsAPI interface {
	Handler(c *gin.Context) error
}

type BankStatsHandler struct {
	builder.Builder
}

func (h *BankStatsHandler) InitFromBuilder(builder builder.Builder) *BankStatsHandler {
	h.Builder = builder
	return h
}

func (h *BankStatsHandler) Handler(context *gin.Context) error {
	userObj, _ := context.Get("user")
	user, _ := userObj.(*models.User)

	bankId := context.Param("bankId")
	if bankId == "" {
		return fmt.Errorf("bankId is required")
	}

	bankAccount, err := h.GetRepository().GetBankAccountById(bankId, user.ID)
	if err != nil {
		return fmt.Errorf("failed to get bank account: %v", err)
	}

	lastTransaction, err := h.GetRepository().GetLastTransactionFromUserBank(user.ID, string(bankId))
	if err != nil {
		return err
	}

	var lastTransactionDate *time.Time
	var daysOutdated *int

	if lastTransaction != nil {
		lastTransactionDate = &lastTransaction.Date
		days := int(time.Since(lastTransaction.Date).Hours() / 24)
		daysOutdated = &days
	}

	transactionCount, err := h.GetRepository().GetTransactionCountFromUserBank(user.ID, string(bankId))
	if err != nil {
		return err
	}

	bankStats := models.BankAccountStat{
		BankAccount:      *bankAccount,
		TransactionCount: transactionCount,
		LastTransaction:  lastTransactionDate,
		DaysOutdated:     daysOutdated,
	}

	context.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   bankStats,
	})
	return nil
}
