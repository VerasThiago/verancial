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

	buildermocks "github.com/verasthiago/verancial/app-integration-worker/pkg/builder/mocks"
	generatormocks "github.com/verasthiago/verancial/app-integration-worker/pkg/generators/mocks"
	apptypes "github.com/verasthiago/verancial/app-integration-worker/pkg/types"
	"github.com/verasthiago/verancial/shared/constants"
	"github.com/verasthiago/verancial/shared/models"
	repositorymocks "github.com/verasthiago/verancial/shared/repository/mocks"
	"github.com/verasthiago/verancial/shared/types"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type handlerTestMocks struct {
	repository *repositorymocks.MockRepository
	factory    *generatormocks.MockAppReportGeneratorFactory
	generator  *generatormocks.MockAppReport
	builder    *buildermocks.MockBuilder
	handler    *AppIntegrationHandler
}

func newHandlerTestMocks(ctrl *gomock.Controller) *handlerTestMocks {
	repository := repositorymocks.NewMockRepository(ctrl)
	factory := generatormocks.NewMockAppReportGeneratorFactory(ctrl)
	generator := generatormocks.NewMockAppReport(ctrl)
	mockBuilder := buildermocks.NewMockBuilder(ctrl)

	handler := new(AppIntegrationHandler).InitFromBuilder(mockBuilder)

	return &handlerTestMocks{
		repository: repository,
		factory:    factory,
		generator:  generator,
		builder:    mockBuilder,
		handler:    handler,
	}
}

func testUser() *models.User {
	return &models.User{
		ID: "user-1",
		FinancialAppCredentials: models.FinancialAppCredentialsMap{
			constants.BudgetBakers: &models.FinancialAppCredentials{
				Login:    "user@example.com",
				Password: "secret",
			},
		},
	}
}

func testPayload() types.AppIntegrationQueuePayload {
	return types.AppIntegrationQueuePayload{
		UserId:              "user-1",
		AppID:               constants.BudgetBakers,
		BankId:              "bank-1",
		LastTransactionDate: "January 2 2024",
	}
}

// --- Execute: happy path and orchestration order ---

func TestExecute_HappyPath_CallsInOrderWithCorrectArgs(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := newHandlerTestMocks(ctrl)

	payload := testPayload()
	user := testUser()
	lastTransactionTime := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
	transactions := []*models.Transaction{
		{ID: "txn-1", UserId: user.ID, Amount: 10.5},
	}
	appReport := apptypes.AppReport{
		{"Date", "Amount", "Note", "Payee", "Currency"},
	}

	gomock.InOrder(
		m.repository.EXPECT().GetUserByID(payload.UserId).Return(user, nil),
		m.builder.EXPECT().GetAppReportGeneratorFactory().Return(m.factory),
		m.factory.EXPECT().GetAppReportGenerator(payload.AppID).Return(m.generator, nil),
		m.generator.EXPECT().
			GetLastTransaction(user.FinancialAppCredentials[payload.AppID], payload.BankId, payload.LastTransactionDate).
			Return(lastTransactionTime, nil),
		m.repository.EXPECT().
			GetAllTransactionsFromUserBankAfterDate(user.ID, payload.BankId, lastTransactionTime).
			Return(transactions, nil),
		m.generator.EXPECT().Generate(transactions).Return(appReport, nil),
		m.generator.EXPECT().Submit(user, appReport).Return(nil),
	)

	m.builder.EXPECT().GetRepository().Return(m.repository).AnyTimes()

	err := m.handler.Execute(payload)

	assert.NoError(t, err)
}

func TestExecute_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := newHandlerTestMocks(ctrl)

	payload := testPayload()
	expectedErr := errors.New("user not found")

	m.builder.EXPECT().GetRepository().Return(m.repository).AnyTimes()
	m.repository.EXPECT().GetUserByID(payload.UserId).Return(nil, expectedErr)

	// None of these should be called.
	m.builder.EXPECT().GetAppReportGeneratorFactory().Times(0)
	m.factory.EXPECT().GetAppReportGenerator(gomock.Any()).Times(0)

	err := m.handler.Execute(payload)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestExecute_UnsupportedApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := newHandlerTestMocks(ctrl)

	payload := testPayload()
	user := testUser()
	expectedErr := errors.New("app not supported")

	m.builder.EXPECT().GetRepository().Return(m.repository).AnyTimes()
	m.repository.EXPECT().GetUserByID(payload.UserId).Return(user, nil)
	m.builder.EXPECT().GetAppReportGeneratorFactory().Return(m.factory)
	m.factory.EXPECT().GetAppReportGenerator(payload.AppID).Return(nil, expectedErr)

	err := m.handler.Execute(payload)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestExecute_GetLastTransactionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := newHandlerTestMocks(ctrl)

	payload := testPayload()
	user := testUser()
	expectedErr := errors.New("invalid date")

	m.builder.EXPECT().GetRepository().Return(m.repository).AnyTimes()
	m.repository.EXPECT().GetUserByID(payload.UserId).Return(user, nil)
	m.builder.EXPECT().GetAppReportGeneratorFactory().Return(m.factory)
	m.factory.EXPECT().GetAppReportGenerator(payload.AppID).Return(m.generator, nil)
	m.generator.EXPECT().
		GetLastTransaction(user.FinancialAppCredentials[payload.AppID], payload.BankId, payload.LastTransactionDate).
		Return(time.Time{}, expectedErr)

	// Should not proceed to fetch transactions.
	m.repository.EXPECT().GetAllTransactionsFromUserBankAfterDate(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	err := m.handler.Execute(payload)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestExecute_GetAllTransactionsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := newHandlerTestMocks(ctrl)

	payload := testPayload()
	user := testUser()
	lastTransactionTime := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
	expectedErr := errors.New("db error")

	m.builder.EXPECT().GetRepository().Return(m.repository).AnyTimes()
	m.repository.EXPECT().GetUserByID(payload.UserId).Return(user, nil)
	m.builder.EXPECT().GetAppReportGeneratorFactory().Return(m.factory)
	m.factory.EXPECT().GetAppReportGenerator(payload.AppID).Return(m.generator, nil)
	m.generator.EXPECT().
		GetLastTransaction(user.FinancialAppCredentials[payload.AppID], payload.BankId, payload.LastTransactionDate).
		Return(lastTransactionTime, nil)
	m.repository.EXPECT().
		GetAllTransactionsFromUserBankAfterDate(user.ID, payload.BankId, lastTransactionTime).
		Return(nil, expectedErr)

	m.generator.EXPECT().Generate(gomock.Any()).Times(0)

	err := m.handler.Execute(payload)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestExecute_GenerateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := newHandlerTestMocks(ctrl)

	payload := testPayload()
	user := testUser()
	lastTransactionTime := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
	transactions := []*models.Transaction{{ID: "txn-1"}}
	expectedErr := errors.New("generate failed")

	m.builder.EXPECT().GetRepository().Return(m.repository).AnyTimes()
	m.repository.EXPECT().GetUserByID(payload.UserId).Return(user, nil)
	m.builder.EXPECT().GetAppReportGeneratorFactory().Return(m.factory)
	m.factory.EXPECT().GetAppReportGenerator(payload.AppID).Return(m.generator, nil)
	m.generator.EXPECT().
		GetLastTransaction(user.FinancialAppCredentials[payload.AppID], payload.BankId, payload.LastTransactionDate).
		Return(lastTransactionTime, nil)
	m.repository.EXPECT().
		GetAllTransactionsFromUserBankAfterDate(user.ID, payload.BankId, lastTransactionTime).
		Return(transactions, nil)
	m.generator.EXPECT().Generate(transactions).Return(nil, expectedErr)

	m.generator.EXPECT().Submit(gomock.Any(), gomock.Any()).Times(0)

	err := m.handler.Execute(payload)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestExecute_SubmitError(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := newHandlerTestMocks(ctrl)

	payload := testPayload()
	user := testUser()
	lastTransactionTime := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
	transactions := []*models.Transaction{{ID: "txn-1"}}
	appReport := apptypes.AppReport{{"Date", "Amount", "Note", "Payee", "Currency"}}
	expectedErr := errors.New("submit failed")

	m.builder.EXPECT().GetRepository().Return(m.repository).AnyTimes()
	m.repository.EXPECT().GetUserByID(payload.UserId).Return(user, nil)
	m.builder.EXPECT().GetAppReportGeneratorFactory().Return(m.factory)
	m.factory.EXPECT().GetAppReportGenerator(payload.AppID).Return(m.generator, nil)
	m.generator.EXPECT().
		GetLastTransaction(user.FinancialAppCredentials[payload.AppID], payload.BankId, payload.LastTransactionDate).
		Return(lastTransactionTime, nil)
	m.repository.EXPECT().
		GetAllTransactionsFromUserBankAfterDate(user.ID, payload.BankId, lastTransactionTime).
		Return(transactions, nil)
	m.generator.EXPECT().Generate(transactions).Return(appReport, nil)
	m.generator.EXPECT().Submit(user, appReport).Return(expectedErr)

	err := m.handler.Execute(payload)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

// --- HandlerSync ---

func TestHandlerSync_HappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := newHandlerTestMocks(ctrl)

	payload := testPayload()
	user := testUser()
	lastTransactionTime := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
	transactions := []*models.Transaction{{ID: "txn-1"}}
	appReport := apptypes.AppReport{{"Date", "Amount", "Note", "Payee", "Currency"}}

	m.builder.EXPECT().GetRepository().Return(m.repository).AnyTimes()
	m.repository.EXPECT().GetUserByID(payload.UserId).Return(user, nil)
	m.builder.EXPECT().GetAppReportGeneratorFactory().Return(m.factory)
	m.factory.EXPECT().GetAppReportGenerator(payload.AppID).Return(m.generator, nil)
	m.generator.EXPECT().
		GetLastTransaction(user.FinancialAppCredentials[payload.AppID], payload.BankId, payload.LastTransactionDate).
		Return(lastTransactionTime, nil)
	m.repository.EXPECT().
		GetAllTransactionsFromUserBankAfterDate(user.ID, payload.BankId, lastTransactionTime).
		Return(transactions, nil)
	m.generator.EXPECT().Generate(transactions).Return(appReport, nil)
	m.generator.EXPECT().Submit(user, appReport).Return(nil)

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	ginCtx, _ := newGinTestContext(http.MethodPost, "/aiw/v0/process_app_report", body)

	err = m.handler.HandlerSync(ginCtx)

	assert.NoError(t, err)
}

func TestHandlerSync_MalformedJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := newHandlerTestMocks(ctrl)

	// No repository/generator calls expected since binding fails first.
	ginCtx, _ := newGinTestContext(http.MethodPost, "/aiw/v0/process_app_report", []byte("{not-valid-json"))

	err := m.handler.HandlerSync(ginCtx)

	assert.Error(t, err)
}

func TestHandlerSync_PropagatesExecuteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := newHandlerTestMocks(ctrl)

	payload := testPayload()
	expectedErr := errors.New("user not found")

	m.builder.EXPECT().GetRepository().Return(m.repository).AnyTimes()
	m.repository.EXPECT().GetUserByID(payload.UserId).Return(nil, expectedErr)

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	ginCtx, _ := newGinTestContext(http.MethodPost, "/aiw/v0/process_app_report", body)

	err = m.handler.HandlerSync(ginCtx)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func newGinTestContext(method, path string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ginCtx.Request = req
	return ginCtx, recorder
}

// --- HandlerAsync ---

func TestHandlerAsync_HappyPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := newHandlerTestMocks(ctrl)

	payload := testPayload()
	user := testUser()
	lastTransactionTime := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
	transactions := []*models.Transaction{{ID: "txn-1"}}
	appReport := apptypes.AppReport{{"Date", "Amount", "Note", "Payee", "Currency"}}

	m.builder.EXPECT().GetRepository().Return(m.repository).AnyTimes()
	m.repository.EXPECT().GetUserByID(payload.UserId).Return(user, nil)
	m.builder.EXPECT().GetAppReportGeneratorFactory().Return(m.factory)
	m.factory.EXPECT().GetAppReportGenerator(payload.AppID).Return(m.generator, nil)
	m.generator.EXPECT().
		GetLastTransaction(user.FinancialAppCredentials[payload.AppID], payload.BankId, payload.LastTransactionDate).
		Return(lastTransactionTime, nil)
	m.repository.EXPECT().
		GetAllTransactionsFromUserBankAfterDate(user.ID, payload.BankId, lastTransactionTime).
		Return(transactions, nil)
	m.generator.EXPECT().Generate(transactions).Return(appReport, nil)
	m.generator.EXPECT().Submit(user, appReport).Return(nil)

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(types.PatternAppIntegration, payloadBytes)

	err = m.handler.HandlerAsync(context.Background(), task)

	assert.NoError(t, err)
}

func TestHandlerAsync_MalformedPayload(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := newHandlerTestMocks(ctrl)

	task := asynq.NewTask(types.PatternAppIntegration, []byte("{not-valid-json"))

	err := m.handler.HandlerAsync(context.Background(), task)

	assert.Error(t, err)
}

func TestHandlerAsync_PropagatesExecuteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := newHandlerTestMocks(ctrl)

	payload := testPayload()
	expectedErr := errors.New("boom")

	m.builder.EXPECT().GetRepository().Return(m.repository).AnyTimes()
	m.repository.EXPECT().GetUserByID(payload.UserId).Return(nil, expectedErr)

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(types.PatternAppIntegration, payloadBytes)

	err = m.handler.HandlerAsync(context.Background(), task)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestInitFromBuilder_SetsBuilder(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockBuilder := buildermocks.NewMockBuilder(ctrl)

	handler := new(AppIntegrationHandler).InitFromBuilder(mockBuilder)

	assert.Same(t, mockBuilder, handler.Builder)
}
