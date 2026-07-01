package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	buildermocks "github.com/verasthiago/verancial/data-process-worker/pkg/builder/mocks"
	reportmocks "github.com/verasthiago/verancial/data-process-worker/pkg/report/mocks"
	"github.com/verasthiago/verancial/shared/constants"
	"github.com/verasthiago/verancial/shared/models"
	repositorymocks "github.com/verasthiago/verancial/shared/repository/mocks"
	"github.com/verasthiago/verancial/shared/types"
)

func newTestHandler(t *testing.T) (*CreateReportHandler, *buildermocks.MockBuilder, *reportmocks.MockReportProcessorFactory, *reportmocks.MockBankReport, *repositorymocks.MockRepository) {
	t.Helper()
	ctrl := gomock.NewController(t)

	mockBuilder := buildermocks.NewMockBuilder(ctrl)
	mockFactory := reportmocks.NewMockReportProcessorFactory(ctrl)
	mockBankReport := reportmocks.NewMockBankReport(ctrl)
	mockRepository := repositorymocks.NewMockRepository(ctrl)

	handler := &CreateReportHandler{}
	handler.InitFromBuilder(mockBuilder)

	return handler, mockBuilder, mockFactory, mockBankReport, mockRepository
}

func samplePayload() types.ReportProcessQueuePayload {
	return types.ReportProcessQueuePayload{
		UserId:   "user-1",
		BankId:   string(constants.ScotiaBank),
		FilePath: "/tmp/report.csv",
	}
}

func TestCreateReportHandler_Execute(t *testing.T) {
	t.Run("happy path calls factory, load, repository lookup, process, fingerprint and save in order", func(t *testing.T) {
		handler, mockBuilder, mockFactory, mockBankReport, mockRepository := newTestHandler(t)
		payload := samplePayload()

		lastDbTransaction := &models.Transaction{Date: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{"raw-row-1"}
		processedTransactions := []*models.Transaction{
			{UserId: "user-1", Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Amount: -10, Description: "desc"},
		}

		gomock.InOrder(
			mockBuilder.EXPECT().GetReportProcessorFactory().Return(mockFactory),
			mockFactory.EXPECT().GetReportProcessor(constants.BankId(payload.BankId)).Return(mockBankReport, nil),
			mockBankReport.EXPECT().LoadFromCSV(payload.FilePath).Return(bankTransactions, nil),
			mockBuilder.EXPECT().GetRepository().Return(mockRepository),
			mockRepository.EXPECT().GetLastTransactionFromUserBank(payload.UserId, payload.BankId).Return(lastDbTransaction, nil),
			mockBankReport.EXPECT().Process(bankTransactions, &payload, lastDbTransaction).Return(processedTransactions, nil),
			mockBuilder.EXPECT().GetRepository().Return(mockRepository),
			mockRepository.EXPECT().CreateUniqueTransactionInBatches(processedTransactions).Return(nil),
		)

		err := handler.Execute(payload)

		require.NoError(t, err)
		// SetFingerprint should have populated the fingerprint before saving.
		assert.NotEmpty(t, processedTransactions[0].Fingerprint)
	})

	t.Run("unsupported bank id: factory error is returned and no further calls happen", func(t *testing.T) {
		handler, mockBuilder, mockFactory, _, _ := newTestHandler(t)
		payload := samplePayload()
		payload.BankId = "unsupported-bank"

		factoryErr := errors.New("bank not supported")

		mockBuilder.EXPECT().GetReportProcessorFactory().Return(mockFactory)
		mockFactory.EXPECT().GetReportProcessor(constants.BankId(payload.BankId)).Return(nil, factoryErr)

		err := handler.Execute(payload)

		require.Error(t, err)
		assert.Equal(t, factoryErr, err)
	})

	t.Run("CSV load error is propagated and stops before repository lookup", func(t *testing.T) {
		handler, mockBuilder, mockFactory, mockBankReport, _ := newTestHandler(t)
		payload := samplePayload()

		loadErr := errors.New("failed to open file")

		mockBuilder.EXPECT().GetReportProcessorFactory().Return(mockFactory)
		mockFactory.EXPECT().GetReportProcessor(constants.BankId(payload.BankId)).Return(mockBankReport, nil)
		mockBankReport.EXPECT().LoadFromCSV(payload.FilePath).Return(nil, loadErr)

		err := handler.Execute(payload)

		require.Error(t, err)
		assert.Equal(t, loadErr, err)
	})

	t.Run("repository GetLastTransactionFromUserBank error is propagated", func(t *testing.T) {
		handler, mockBuilder, mockFactory, mockBankReport, mockRepository := newTestHandler(t)
		payload := samplePayload()

		repoErr := errors.New("db connection lost")
		bankTransactions := []interface{}{"raw-row-1"}

		mockBuilder.EXPECT().GetReportProcessorFactory().Return(mockFactory)
		mockFactory.EXPECT().GetReportProcessor(constants.BankId(payload.BankId)).Return(mockBankReport, nil)
		mockBankReport.EXPECT().LoadFromCSV(payload.FilePath).Return(bankTransactions, nil)
		mockBuilder.EXPECT().GetRepository().Return(mockRepository)
		mockRepository.EXPECT().GetLastTransactionFromUserBank(payload.UserId, payload.BankId).Return(nil, repoErr)

		err := handler.Execute(payload)

		require.Error(t, err)
		assert.Equal(t, repoErr, err)
	})

	t.Run("nil lastDbTransaction defaults to year-2000 workaround before calling Process", func(t *testing.T) {
		handler, mockBuilder, mockFactory, mockBankReport, mockRepository := newTestHandler(t)
		payload := samplePayload()
		bankTransactions := []interface{}{"raw-row-1"}

		expectedDefault := &models.Transaction{Date: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)}

		mockBuilder.EXPECT().GetReportProcessorFactory().Return(mockFactory)
		mockFactory.EXPECT().GetReportProcessor(constants.BankId(payload.BankId)).Return(mockBankReport, nil)
		mockBankReport.EXPECT().LoadFromCSV(payload.FilePath).Return(bankTransactions, nil)
		mockBuilder.EXPECT().GetRepository().Return(mockRepository)
		mockRepository.EXPECT().GetLastTransactionFromUserBank(payload.UserId, payload.BankId).Return(nil, nil)
		mockBankReport.EXPECT().Process(bankTransactions, &payload, expectedDefault).Return(nil, nil)
		mockBuilder.EXPECT().GetRepository().Return(mockRepository)
		mockRepository.EXPECT().CreateUniqueTransactionInBatches(nil).Return(nil)

		err := handler.Execute(payload)

		require.NoError(t, err)
	})

	t.Run("Process error is propagated and save is never called", func(t *testing.T) {
		handler, mockBuilder, mockFactory, mockBankReport, mockRepository := newTestHandler(t)
		payload := samplePayload()
		bankTransactions := []interface{}{"raw-row-1"}
		lastDbTransaction := &models.Transaction{Date: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)}
		processErr := errors.New("bad transaction type")

		mockBuilder.EXPECT().GetReportProcessorFactory().Return(mockFactory)
		mockFactory.EXPECT().GetReportProcessor(constants.BankId(payload.BankId)).Return(mockBankReport, nil)
		mockBankReport.EXPECT().LoadFromCSV(payload.FilePath).Return(bankTransactions, nil)
		mockBuilder.EXPECT().GetRepository().Return(mockRepository)
		mockRepository.EXPECT().GetLastTransactionFromUserBank(payload.UserId, payload.BankId).Return(lastDbTransaction, nil)
		mockBankReport.EXPECT().Process(bankTransactions, &payload, lastDbTransaction).Return(nil, processErr)

		err := handler.Execute(payload)

		require.Error(t, err)
		assert.Equal(t, processErr, err)
	})

	t.Run("repository save error is propagated", func(t *testing.T) {
		handler, mockBuilder, mockFactory, mockBankReport, mockRepository := newTestHandler(t)
		payload := samplePayload()
		bankTransactions := []interface{}{"raw-row-1"}
		lastDbTransaction := &models.Transaction{Date: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)}
		processedTransactions := []*models.Transaction{
			{UserId: "user-1", Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Amount: -10, Description: "desc"},
		}
		saveErr := errors.New("unique constraint violated")

		mockBuilder.EXPECT().GetReportProcessorFactory().Return(mockFactory)
		mockFactory.EXPECT().GetReportProcessor(constants.BankId(payload.BankId)).Return(mockBankReport, nil)
		mockBankReport.EXPECT().LoadFromCSV(payload.FilePath).Return(bankTransactions, nil)
		mockBuilder.EXPECT().GetRepository().Return(mockRepository)
		mockRepository.EXPECT().GetLastTransactionFromUserBank(payload.UserId, payload.BankId).Return(lastDbTransaction, nil)
		mockBankReport.EXPECT().Process(bankTransactions, &payload, lastDbTransaction).Return(processedTransactions, nil)
		mockBuilder.EXPECT().GetRepository().Return(mockRepository)
		mockRepository.EXPECT().CreateUniqueTransactionInBatches(processedTransactions).Return(saveErr)

		err := handler.Execute(payload)

		require.Error(t, err)
		assert.Equal(t, saveErr, err)
	})
}

func TestCreateReportHandler_HandlerSync(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("parses JSON body and delegates to Execute", func(t *testing.T) {
		handler, mockBuilder, mockFactory, mockBankReport, mockRepository := newTestHandler(t)
		payload := samplePayload()

		lastDbTransaction := &models.Transaction{Date: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{"row"}

		mockBuilder.EXPECT().GetReportProcessorFactory().Return(mockFactory)
		mockFactory.EXPECT().GetReportProcessor(constants.BankId(payload.BankId)).Return(mockBankReport, nil)
		mockBankReport.EXPECT().LoadFromCSV(payload.FilePath).Return(bankTransactions, nil)
		mockBuilder.EXPECT().GetRepository().Return(mockRepository)
		mockRepository.EXPECT().GetLastTransactionFromUserBank(payload.UserId, payload.BankId).Return(lastDbTransaction, nil)
		mockBankReport.EXPECT().Process(bankTransactions, gomock.Any(), lastDbTransaction).Return(nil, nil)
		mockBuilder.EXPECT().GetRepository().Return(mockRepository)
		mockRepository.EXPECT().CreateUniqueTransactionInBatches(nil).Return(nil)

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodPost, "/report", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		ginCtx.Request = req

		err = handler.HandlerSync(ginCtx)

		require.NoError(t, err)
	})

	t.Run("invalid JSON body returns bind error without calling builder", func(t *testing.T) {
		handler, _, _, _, _ := newTestHandler(t)

		w := httptest.NewRecorder()
		ginCtx, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodPost, "/report", bytes.NewReader([]byte("not-json")))
		req.Header.Set("Content-Type", "application/json")
		ginCtx.Request = req

		err := handler.HandlerSync(ginCtx)

		require.Error(t, err)
	})
}

func TestCreateReportHandler_HandlerAsync(t *testing.T) {
	t.Run("parses task payload and delegates to Execute", func(t *testing.T) {
		handler, mockBuilder, mockFactory, mockBankReport, mockRepository := newTestHandler(t)
		payload := samplePayload()

		lastDbTransaction := &models.Transaction{Date: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)}
		bankTransactions := []interface{}{"row"}

		mockBuilder.EXPECT().GetReportProcessorFactory().Return(mockFactory)
		mockFactory.EXPECT().GetReportProcessor(constants.BankId(payload.BankId)).Return(mockBankReport, nil)
		mockBankReport.EXPECT().LoadFromCSV(payload.FilePath).Return(bankTransactions, nil)
		mockBuilder.EXPECT().GetRepository().Return(mockRepository)
		mockRepository.EXPECT().GetLastTransactionFromUserBank(payload.UserId, payload.BankId).Return(lastDbTransaction, nil)
		mockBankReport.EXPECT().Process(bankTransactions, gomock.Any(), lastDbTransaction).Return(nil, nil)
		mockBuilder.EXPECT().GetRepository().Return(mockRepository)
		mockRepository.EXPECT().CreateUniqueTransactionInBatches(nil).Return(nil)

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		task := asynq.NewTask("report:create", body)

		err = handler.HandlerAsync(context.Background(), task)

		require.NoError(t, err)
	})

	t.Run("invalid task payload returns unmarshal error without calling builder", func(t *testing.T) {
		handler, _, _, _, _ := newTestHandler(t)

		task := asynq.NewTask("report:create", []byte("not-json"))

		err := handler.HandlerAsync(context.Background(), task)

		require.Error(t, err)
	})
}
