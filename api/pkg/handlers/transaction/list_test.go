package transaction

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	buildermocks "github.com/verasthiago/verancial/api/pkg/builder/mocks"
	"github.com/verasthiago/verancial/api/pkg/testutil"
	"github.com/verasthiago/verancial/shared/models"
	repomocks "github.com/verasthiago/verancial/shared/repository/mocks"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validBankID = "3fa85f64-5717-4562-b3fc-2c963f66afa6"

func newListContext(target string, bankID string) (*gin.Context, *httptest.ResponseRecorder) {
	c, w := testutil.NewGinContext("GET", target, nil, "")
	c.Params = gin.Params{{Key: "bankId", Value: bankID}}
	return c, w
}

func TestListTransactionsHandler_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success with default pagination and filters", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockRepo := repomocks.NewMockRepository(ctrl)

		userID := "user-1"
		txs := []*models.Transaction{{ID: "tx-1"}, {ID: "tx-2"}}

		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockRepo.EXPECT().GetTransactions(userID, validBankID, 50, 0, &models.TransactionFilter{Uncategorized: false}).Return(txs, nil)

		c, w := newListContext("/api/v0/transaction/list/"+validBankID, validBankID)
		testutil.SetAuthenticatedUser(c, userID, false)

		handler := new(ListTransactionsHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.NoError(t, err)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("pagination params are respected", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockRepo := repomocks.NewMockRepository(ctrl)

		userID := "user-1"
		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		// page=3, pageSize=10 => offset = (3-1)*10 = 20
		mockRepo.EXPECT().GetTransactions(userID, validBankID, 10, 20, &models.TransactionFilter{Uncategorized: false}).Return([]*models.Transaction{}, nil)

		c, _ := newListContext("/api/v0/transaction/list/"+validBankID+"?page=3&pageSize=10", validBankID)
		testutil.SetAuthenticatedUser(c, userID, false)

		handler := new(ListTransactionsHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.NoError(t, err)
	})

	t.Run("uncategorized filter is respected", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockRepo := repomocks.NewMockRepository(ctrl)

		userID := "user-1"
		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockRepo.EXPECT().GetTransactions(userID, validBankID, 50, 0, &models.TransactionFilter{Uncategorized: true}).Return([]*models.Transaction{}, nil)

		c, _ := newListContext("/api/v0/transaction/list/"+validBankID+"?uncategorized=true", validBankID)
		testutil.SetAuthenticatedUser(c, userID, false)

		handler := new(ListTransactionsHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.NoError(t, err)
	})

	t.Run("response body contains transactions", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockRepo := repomocks.NewMockRepository(ctrl)

		userID := "user-1"
		txs := []*models.Transaction{{ID: "tx-1", Payee: "Coffee Shop"}}
		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockRepo.EXPECT().GetTransactions(userID, validBankID, 50, 0, gomock.Any()).Return(txs, nil)

		c, w := newListContext("/api/v0/transaction/list/"+validBankID, validBankID)
		testutil.SetAuthenticatedUser(c, userID, false)

		handler := new(ListTransactionsHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)
		require.NoError(t, err)

		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, "success", body["status"])
		data := body["data"].([]interface{})
		require.Len(t, data, 1)
		assert.Equal(t, "Coffee Shop", data[0].(map[string]interface{})["payee"])
	})

	t.Run("no authenticated user returns unauthorized error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		c, _ := newListContext("/api/v0/transaction/list/"+validBankID, validBankID)

		handler := new(ListTransactionsHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("invalid bankId (not a UUID) returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		c, _ := newListContext("/api/v0/transaction/list/not-a-uuid", "not-a-uuid")
		testutil.SetAuthenticatedUser(c, "user-1", false)

		handler := new(ListTransactionsHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid bank ID")
	})

	t.Run("missing bankId param returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		c, _ := testutil.NewGinContext("GET", "/api/v0/transaction/list/", nil, "")
		testutil.SetAuthenticatedUser(c, "user-1", false)
		// no bankId param set at all

		handler := new(ListTransactionsHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.Error(t, err)
	})

	t.Run("page below minimum returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		c, _ := newListContext("/api/v0/transaction/list/"+validBankID+"?page=0", validBankID)
		testutil.SetAuthenticatedUser(c, "user-1", false)

		handler := new(ListTransactionsHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid query parameters")
	})

	t.Run("pageSize above maximum returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		c, _ := newListContext("/api/v0/transaction/list/"+validBankID+"?pageSize=101", validBankID)
		testutil.SetAuthenticatedUser(c, "user-1", false)

		handler := new(ListTransactionsHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid query parameters")
	})

	t.Run("repository error is propagated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockRepo := repomocks.NewMockRepository(ctrl)

		userID := "user-1"
		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockRepo.EXPECT().GetTransactions(userID, validBankID, 50, 0, gomock.Any()).Return(nil, fmt.Errorf("db exploded"))

		c, _ := newListContext("/api/v0/transaction/list/"+validBankID, validBankID)
		testutil.SetAuthenticatedUser(c, userID, false)

		handler := new(ListTransactionsHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get transactions")
	})
}
