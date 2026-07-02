package handlers

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	buildermocks "github.com/verasthiago/verancial/api/pkg/builder/mocks"
	"github.com/verasthiago/verancial/api/pkg/testutil"
	"github.com/verasthiago/verancial/shared/models"
	repomocks "github.com/verasthiago/verancial/shared/repository/mocks"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBankStatsHandler_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns bank stats as JSON", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockRepo := repomocks.NewMockRepository(ctrl)

		userID := "user-1"
		bankID := "bank-1"
		lastTxTime := time.Now().Add(-48 * time.Hour)

		bankAccount := &models.BankAccount{ID: bankID, Name: "Chase"}
		lastTx := &models.Transaction{ID: "tx-1", Date: lastTxTime}

		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockRepo.EXPECT().GetBankAccountById(bankID, userID).Return(bankAccount, nil)
		mockRepo.EXPECT().GetLastTransactionFromUserBank(userID, bankID).Return(lastTx, nil)
		mockRepo.EXPECT().GetTransactionCountFromUserBank(userID, bankID).Return(5, nil)

		c, w := testutil.NewGinContext("GET", "/api/v0/bank/"+bankID, nil, "")
		c.Params = gin.Params{{Key: "bankId", Value: bankID}}
		testutil.SetAuthenticatedUser(c, userID, false)

		handler := new(BankStatsHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.NoError(t, err)
		assert.Equal(t, 200, w.Code)

		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, "success", body["status"])
		data := body["data"].(map[string]interface{})
		assert.Equal(t, float64(5), data["transaction_count"])
		assert.NotNil(t, data["days_outdated"])
	})

	t.Run("success with no prior transactions leaves last transaction nil", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockRepo := repomocks.NewMockRepository(ctrl)

		userID := "user-1"
		bankID := "bank-1"
		bankAccount := &models.BankAccount{ID: bankID, Name: "Chase"}

		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockRepo.EXPECT().GetBankAccountById(bankID, userID).Return(bankAccount, nil)
		mockRepo.EXPECT().GetLastTransactionFromUserBank(userID, bankID).Return(nil, nil)
		mockRepo.EXPECT().GetTransactionCountFromUserBank(userID, bankID).Return(0, nil)

		c, w := testutil.NewGinContext("GET", "/api/v0/bank/"+bankID, nil, "")
		c.Params = gin.Params{{Key: "bankId", Value: bankID}}
		testutil.SetAuthenticatedUser(c, userID, false)

		handler := new(BankStatsHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.NoError(t, err)
		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		data := body["data"].(map[string]interface{})
		assert.Nil(t, data["last_transaction"])
		assert.Nil(t, data["days_outdated"])
	})

	t.Run("no authenticated user returns unauthorized error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		c, _ := testutil.NewGinContext("GET", "/api/v0/bank/bank-1", nil, "")
		c.Params = gin.Params{{Key: "bankId", Value: "bank-1"}}
		// user not set on context

		handler := new(BankStatsHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("missing bankId param returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		c, _ := testutil.NewGinContext("GET", "/api/v0/bank/", nil, "")
		testutil.SetAuthenticatedUser(c, "user-1", false)
		// bankId param intentionally not set

		handler := new(BankStatsHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "bankId is required")
	})

	t.Run("GetBankAccountById error is propagated (not found path)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockRepo := repomocks.NewMockRepository(ctrl)

		userID := "user-1"
		bankID := "missing-bank"

		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockRepo.EXPECT().GetBankAccountById(bankID, userID).Return(nil, fmt.Errorf("record not found"))

		c, _ := testutil.NewGinContext("GET", "/api/v0/bank/"+bankID, nil, "")
		c.Params = gin.Params{{Key: "bankId", Value: bankID}}
		testutil.SetAuthenticatedUser(c, userID, false)

		handler := new(BankStatsHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get bank account")
	})

	t.Run("GetLastTransactionFromUserBank error is propagated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockRepo := repomocks.NewMockRepository(ctrl)

		userID := "user-1"
		bankID := "bank-1"
		bankAccount := &models.BankAccount{ID: bankID}

		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockRepo.EXPECT().GetBankAccountById(bankID, userID).Return(bankAccount, nil)
		mockRepo.EXPECT().GetLastTransactionFromUserBank(userID, bankID).Return(nil, fmt.Errorf("db exploded"))

		c, _ := testutil.NewGinContext("GET", "/api/v0/bank/"+bankID, nil, "")
		c.Params = gin.Params{{Key: "bankId", Value: bankID}}
		testutil.SetAuthenticatedUser(c, userID, false)

		handler := new(BankStatsHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "db exploded")
	})

	t.Run("GetTransactionCountFromUserBank error is propagated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockRepo := repomocks.NewMockRepository(ctrl)

		userID := "user-1"
		bankID := "bank-1"
		bankAccount := &models.BankAccount{ID: bankID}

		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockRepo.EXPECT().GetBankAccountById(bankID, userID).Return(bankAccount, nil)
		mockRepo.EXPECT().GetLastTransactionFromUserBank(userID, bankID).Return(nil, nil)
		mockRepo.EXPECT().GetTransactionCountFromUserBank(userID, bankID).Return(0, fmt.Errorf("count failed"))

		c, _ := testutil.NewGinContext("GET", "/api/v0/bank/"+bankID, nil, "")
		c.Params = gin.Params{{Key: "bankId", Value: bankID}}
		testutil.SetAuthenticatedUser(c, userID, false)

		handler := new(BankStatsHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "count failed")
	})
}
