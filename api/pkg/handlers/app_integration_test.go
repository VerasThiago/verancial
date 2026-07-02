package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	buildermocks "github.com/verasthiago/verancial/api/pkg/builder/mocks"
	"github.com/verasthiago/verancial/api/pkg/testutil"
	"github.com/verasthiago/verancial/shared/auth"
	"github.com/verasthiago/verancial/shared/constants"
	sharedflags "github.com/verasthiago/verancial/shared/flags"
	httpclientmocks "github.com/verasthiago/verancial/shared/httpclient/mocks"
	"github.com/verasthiago/verancial/shared/types"
	taskmocks "github.com/verasthiago/verancial/shared/task/mocks"

	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- IDOR regression test ----------------------------------------------------

// TestAppIntegrationHandler_Handler_IDOR proves UserId always comes from the
// authenticated context user, never from client JSON, even if the client
// tries to spoof it. AppIntegrationRequest.UserId has `json:"-"`, so it can
// never be bound from the body -- confirmed here end-to-end via Handler.
func TestAppIntegrationHandler_Handler_IDOR(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilder := buildermocks.NewMockBuilder(ctrl)
	mockTask := taskmocks.NewMockTask(ctrl)

	realUserID := "real-authenticated-user"
	spoofedUserID := "victim-user-id"

	bodyJSON := fmt.Sprintf(`{
		"UserId": %q,
		"userid": %q,
		"appid": "budgetbakers",
		"bankid": "bank-1",
		"lasttransactiondate": "2024-01-01",
		"asyncprocessing": true
	}`, spoofedUserID, spoofedUserID)

	mockBuilder.EXPECT().GetTask().Return(mockTask).AnyTimes()

	var captured types.AppIntegrationQueuePayload
	mockTask.EXPECT().UpdateAppAsync(gomock.Any()).DoAndReturn(func(p types.AppIntegrationQueuePayload) error {
		captured = p
		return nil
	})

	c, w := testutil.NewGinContext("POST", "/api/v0/app-integration/generate", []byte(bodyJSON), "")
	testutil.SetAuthenticatedUser(c, realUserID, false)

	handler := new(AppIntegrationHandler).InitFromBuilder(mockBuilder)
	err := handler.Handler(c)

	require.NoError(t, err)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, realUserID, captured.UserId)
	assert.NotEqual(t, spoofedUserID, captured.UserId)
}

func TestAppIntegrationHandler_Handler_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("no user in context", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		body := []byte(`{"appid":"budgetbakers","bankid":"bank-1"}`)
		c, _ := testutil.NewGinContext("POST", "/api/v0/app-integration/generate", body, "")

		handler := new(AppIntegrationHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("user with empty ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		body := []byte(`{"appid":"budgetbakers","bankid":"bank-1"}`)
		c, _ := testutil.NewGinContext("POST", "/api/v0/app-integration/generate", body, "")
		c.Set("user", &auth.UserClaims{ID: ""})

		handler := new(AppIntegrationHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("malformed JSON body", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		body := []byte(`{not-json`)
		c, _ := testutil.NewGinContext("POST", "/api/v0/app-integration/generate", body, "")
		testutil.SetAuthenticatedUser(c, "user-1", false)

		handler := new(AppIntegrationHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.Error(t, err)
	})
}

// --- HandlerAsync (MockTask) --------------------------------------------------

func TestAppIntegrationHandler_HandlerAsync(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("calls UpdateAppAsync with the right payload and returns queued", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockTask := taskmocks.NewMockTask(ctrl)
		mockBuilder.EXPECT().GetTask().Return(mockTask)

		expectedPayload := types.AppIntegrationQueuePayload{
			UserId:              "user-1",
			AppID:               constants.AppID("budgetbakers"),
			BankId:              "bank-1",
			LastTransactionDate: "2024-01-01",
		}
		mockTask.EXPECT().UpdateAppAsync(expectedPayload).Return(nil)

		c, w := testutil.NewGinContext("POST", "/api/v0/app-integration/generate", nil, "")

		handler := new(AppIntegrationHandler).InitFromBuilder(mockBuilder)
		req := AppIntegrationRequest{
			UserId:              "user-1",
			AppID:               "budgetbakers",
			BankId:              "bank-1",
			LastTransactionData: "2024-01-01",
		}
		err := handler.HandlerAsync(req, c)

		require.NoError(t, err)
		assert.Equal(t, 200, w.Code)

		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, "queued", body["status"])
	})

	t.Run("task error is propagated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockTask := taskmocks.NewMockTask(ctrl)
		mockBuilder.EXPECT().GetTask().Return(mockTask)
		mockTask.EXPECT().UpdateAppAsync(gomock.Any()).Return(fmt.Errorf("queue down"))

		c, _ := testutil.NewGinContext("POST", "/api/v0/app-integration/generate", nil, "")

		handler := new(AppIntegrationHandler).InitFromBuilder(mockBuilder)
		req := AppIntegrationRequest{UserId: "user-1"}
		err := handler.HandlerAsync(req, c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "queue down")
	})
}

// --- HandlerSync (MockHTTPClient) --------------------------------------------

func TestAppIntegrationHandler_HandlerSync(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 OK from AIW returns success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockHTTPClient := httpclientmocks.NewMockHTTPClient(ctrl)

		mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{AIWHost: "aiw-host", AIWPort: "9100"}).AnyTimes()
		mockBuilder.EXPECT().GetHTTPClient().Return(mockHTTPClient).AnyTimes()

		var gotMethod, gotURL, gotContentType string
		var gotBody []byte

		mockHTTPClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
			gotMethod = req.Method
			gotURL = req.URL.String()
			gotContentType = req.Header.Get("Content-Type")
			gotBody, _ = io.ReadAll(req.Body)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		})

		c, w := testutil.NewGinContext("POST", "/api/v0/app-integration/generate", nil, "")

		handler := new(AppIntegrationHandler).InitFromBuilder(mockBuilder)
		req := AppIntegrationRequest{UserId: "user-1", AppID: "budgetbakers", BankId: "bank-1"}
		err := handler.HandlerSync(req, c)

		require.NoError(t, err)
		assert.Equal(t, 200, w.Code)

		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, "App report generated successfully", body["status"])

		assert.Equal(t, "POST", gotMethod)
		assert.Equal(t, "http://aiw-host:9100/aiw/v0/process_app_report", gotURL)
		assert.Equal(t, "application/json", gotContentType)

		var sentReq types.AppIntegrationQueuePayload
		require.NoError(t, json.Unmarshal(gotBody, &sentReq))
		assert.Equal(t, "user-1", sentReq.UserId, "UserId must reach AIW -- it's tagged json:\"-\" on the local AppIntegrationRequest type (to stop a spoofed UserId in the inbound client body), so the outbound payload must NOT reuse that type or UserId silently vanishes")
		assert.Equal(t, constants.AppID("budgetbakers"), sentReq.AppID)
		assert.Equal(t, "bank-1", sentReq.BankId)
	})

	t.Run("non-200 status from AIW returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockHTTPClient := httpclientmocks.NewMockHTTPClient(ctrl)

		mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{AIWHost: "aiw-host", AIWPort: "9100"}).AnyTimes()
		mockBuilder.EXPECT().GetHTTPClient().Return(mockHTTPClient).AnyTimes()

		mockHTTPClient.EXPECT().Do(gomock.Any()).Return(&http.Response{
			StatusCode: http.StatusBadGateway,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil)

		c, _ := testutil.NewGinContext("POST", "/api/v0/app-integration/generate", nil, "")

		handler := new(AppIntegrationHandler).InitFromBuilder(mockBuilder)
		req := AppIntegrationRequest{UserId: "user-1"}
		err := handler.HandlerSync(req, c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "502")
		assert.Contains(t, err.Error(), "AIW")
	})

	t.Run("network error from HTTPClient is propagated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockHTTPClient := httpclientmocks.NewMockHTTPClient(ctrl)

		mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{AIWHost: "aiw-host", AIWPort: "9100"}).AnyTimes()
		mockBuilder.EXPECT().GetHTTPClient().Return(mockHTTPClient).AnyTimes()

		mockHTTPClient.EXPECT().Do(gomock.Any()).Return(nil, fmt.Errorf("no route to host"))

		c, _ := testutil.NewGinContext("POST", "/api/v0/app-integration/generate", nil, "")

		handler := new(AppIntegrationHandler).InitFromBuilder(mockBuilder)
		req := AppIntegrationRequest{UserId: "user-1"}
		err := handler.HandlerSync(req, c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no route to host")
	})
}

// --- Handler dispatch: async vs sync -----------------------------------------

func TestAppIntegrationHandler_Handler_DispatchesAsyncVsSync(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("asyncprocessing true routes to task queue", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockTask := taskmocks.NewMockTask(ctrl)
		mockBuilder.EXPECT().GetTask().Return(mockTask)
		mockTask.EXPECT().UpdateAppAsync(gomock.Any()).Return(nil)

		body := []byte(`{"appid":"budgetbakers","bankid":"bank-1","asyncprocessing":true}`)
		c, w := testutil.NewGinContext("POST", "/api/v0/app-integration/generate", body, "")
		testutil.SetAuthenticatedUser(c, "user-1", false)

		handler := new(AppIntegrationHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.NoError(t, err)
		var respBody map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respBody))
		assert.Equal(t, "queued", respBody["status"])
	})

	t.Run("asyncprocessing false routes to sync HTTP call", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockHTTPClient := httpclientmocks.NewMockHTTPClient(ctrl)
		mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{AIWHost: "aiw-host", AIWPort: "9100"}).AnyTimes()
		mockBuilder.EXPECT().GetHTTPClient().Return(mockHTTPClient).AnyTimes()
		mockHTTPClient.EXPECT().Do(gomock.Any()).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil)

		body := []byte(`{"appid":"budgetbakers","bankid":"bank-1","asyncprocessing":false}`)
		c, w := testutil.NewGinContext("POST", "/api/v0/app-integration/generate", body, "")
		testutil.SetAuthenticatedUser(c, "user-1", false)

		handler := new(AppIntegrationHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.NoError(t, err)
		var respBody map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &respBody))
		assert.Equal(t, "App report generated successfully", respBody["status"])
	})
}
