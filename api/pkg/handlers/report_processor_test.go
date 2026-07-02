package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	buildermocks "github.com/verasthiago/verancial/api/pkg/builder/mocks"
	"github.com/verasthiago/verancial/api/pkg/testutil"
	"github.com/verasthiago/verancial/shared/auth"
	sharedflags "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/shared/httpclient"
	httpclientmocks "github.com/verasthiago/verancial/shared/httpclient/mocks"
	"github.com/verasthiago/verancial/shared/types"

	"go.uber.org/mock/gomock"

	taskmocks "github.com/verasthiago/verancial/shared/task/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- sanitizeFileName / handleFileUpload -----------------------------------

func TestSanitizeFileName(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"plain filename", "statement.csv", "statement.csv"},
		{"path traversal unix", "../../etc/passwd", "passwd"},
		{"absolute unix path", "/etc/passwd", "passwd"},
		{"nested traversal", "../../../../a/b/../../c.csv", "c.csv"},
		{"only dots", "..", "upload"},
		{"single dot", ".", "upload"},
		{"empty string", "", "upload"},
		{"embedded special chars", "evil;rm -rf$.csv", "evil_rm_-rf_.csv"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeFileName(tc.input)
			assert.Equal(t, tc.want, got)
			// Whatever comes out must never contain a path separator or ".."
			assert.NotContains(t, got, "/")
			assert.NotContains(t, got, `\`)
			assert.NotEqual(t, "..", got)
		})
	}

	// `\` is only a path separator to filepath.Base on Windows; on Linux
	// (where this service actually runs -- Docker/CI) it's just a regular
	// character, so a Windows-style traversal string isn't split into
	// components there. Assert the safety invariant that actually matters
	// (no separator survives, no ".." survives) rather than an exact string
	// that's inherently OS-dependent.
	t.Run("path traversal windows-style backslashes", func(t *testing.T) {
		got := sanitizeFileName(`..\..\..\Windows\System32\config`)
		assert.NotContains(t, got, "/")
		assert.NotContains(t, got, `\`)
		assert.NotEqual(t, "..", got)
		assert.NotEmpty(t, got)
	})
}

func TestReportProcessorHandler_handleFileUpload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid base64 decodes and writes to a sanitized temp path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		handler := new(ReportProcessorHandler).InitFromBuilder(mockBuilder)

		content := "date,amount,payee\n2024-01-01,10.00,Coffee"
		encoded := base64.StdEncoding.EncodeToString([]byte(content))

		req := Request{
			UserId:   "user-abc",
			FileName: "statement.csv",
			FileData: encoded,
		}

		path, err := handler.handleFileUpload(req)
		require.NoError(t, err)
		defer os.Remove(path)

		// Path must live directly inside the OS temp dir (flat, no subdirs).
		tempDir := filepath.Clean(os.TempDir())
		assert.Equal(t, tempDir, filepath.Clean(filepath.Dir(path)))

		writtenData, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, content, string(writtenData))

		assert.Contains(t, filepath.Base(path), "user-abc")
		assert.Contains(t, filepath.Base(path), "statement.csv")
	})

	t.Run("missing FileData errors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		handler := new(ReportProcessorHandler).InitFromBuilder(mockBuilder)

		req := Request{UserId: "user-abc", FileName: "statement.csv"}
		_, err := handler.handleFileUpload(req)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "file data is required")
	})

	t.Run("invalid base64 errors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		handler := new(ReportProcessorHandler).InitFromBuilder(mockBuilder)

		req := Request{UserId: "user-abc", FileName: "statement.csv", FileData: "not-valid-base64!!!"}
		_, err := handler.handleFileUpload(req)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode file data")
	})

	// Path traversal regression test: a malicious FileName must never let
	// the write escape os.TempDir(), regardless of how many "../" segments
	// or absolute-path tricks the client sends.
	t.Run("path traversal in FileName is neutralized", func(t *testing.T) {
		maliciousNames := []string{
			"../../../../etc/passwd",
			`..\..\..\..\Windows\System32\drivers\etc\hosts`,
			"/etc/passwd",
			"....//....//etc/passwd",
			"../../../root/.ssh/authorized_keys",
		}

		for _, malicious := range maliciousNames {
			t.Run(malicious, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				mockBuilder := buildermocks.NewMockBuilder(ctrl)
				handler := new(ReportProcessorHandler).InitFromBuilder(mockBuilder)

				content := "malicious payload"
				encoded := base64.StdEncoding.EncodeToString([]byte(content))

				req := Request{
					UserId:   "victim-user",
					FileName: malicious,
					FileData: encoded,
				}

				path, err := handler.handleFileUpload(req)
				require.NoError(t, err)
				defer os.Remove(path)

				tempDir := os.TempDir()

				// The resulting path must be a direct child of tempDir, i.e.
				// filepath.Dir(path) must equal tempDir (Clean'd) -- no
				// traversal outside of it, no subdirectories created.
				cleanTempDir := filepath.Clean(tempDir)
				assert.Equal(t, cleanTempDir, filepath.Clean(filepath.Dir(path)),
					"file must be written directly inside the temp dir, not escape it")

				rel, err := filepath.Rel(tempDir, path)
				require.NoError(t, err)
				assert.False(t, strings.HasPrefix(rel, ".."), "resolved path escaped tempDir: %s", rel)

				// It must never actually land on a sensitive absolute path.
				assert.NotEqual(t, "/etc/passwd", path)
				assert.NotContains(t, path, filepath.Join("etc", "passwd"))
			})
		}
	})

	t.Run("malicious UserId is also sanitized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		handler := new(ReportProcessorHandler).InitFromBuilder(mockBuilder)

		encoded := base64.StdEncoding.EncodeToString([]byte("data"))
		req := Request{
			UserId:   "../../evil",
			FileName: "report.csv",
			FileData: encoded,
		}

		path, err := handler.handleFileUpload(req)
		require.NoError(t, err)
		defer os.Remove(path)

		tempDir := filepath.Clean(os.TempDir())
		assert.Equal(t, tempDir, filepath.Clean(filepath.Dir(path)))
	})
}

// --- IDOR regression test ----------------------------------------------------

// TestReportProcessorHandler_Handler_IDOR proves that UserId always comes
// from the authenticated JWT/context user, never from client-supplied JSON,
// even when the client attempts to spoof a different user via extra JSON
// fields. Request.UserId has `json:"-"` so it cannot be bound from the body
// at all; this test exercises the full Handler path (JSON bind + context
// user extraction) to prove that end-to-end.
func TestReportProcessorHandler_Handler_IDOR(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilder := buildermocks.NewMockBuilder(ctrl)
	mockTask := taskmocks.NewMockTask(ctrl)

	realUserID := "real-authenticated-user"
	spoofedUserID := "victim-user-id"

	encoded := base64.StdEncoding.EncodeToString([]byte("csv,data"))

	// Client tries to inject a "UserId" field (and its lowercase/camel
	// variants) into the JSON body, hoping the handler will trust it.
	bodyJSON := fmt.Sprintf(`{
		"UserId": %q,
		"userid": %q,
		"filedata": %q,
		"filename": "report.csv",
		"bankid": "bank-1",
		"asyncprocessing": true
	}`, spoofedUserID, spoofedUserID, encoded)

	mockBuilder.EXPECT().GetTask().Return(mockTask).AnyTimes()

	var capturedPayload types.ReportProcessQueuePayload
	mockTask.EXPECT().CreateReportAsync(gomock.Any()).DoAndReturn(func(p types.ReportProcessQueuePayload) error {
		capturedPayload = p
		return nil
	})

	c, w := testutil.NewGinContext("POST", "/api/v0/report/process", []byte(bodyJSON), "")
	testutil.SetAuthenticatedUser(c, realUserID, false)

	handler := new(ReportProcessorHandler).InitFromBuilder(mockBuilder)
	err := handler.Handler(c)
	require.NoError(t, err)
	assert.Equal(t, 200, w.Code)

	defer os.Remove(capturedPayload.FilePath)

	assert.Equal(t, realUserID, capturedPayload.UserId, "UserId must be derived from the JWT/context user, not client JSON")
	assert.NotEqual(t, spoofedUserID, capturedPayload.UserId)
}

func TestReportProcessorHandler_Handler_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("no user in context returns unauthorized error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		body := []byte(`{"filedata":"","filename":"x.csv","bankid":"b"}`)
		c, _ := testutil.NewGinContext("POST", "/api/v0/report/process", body, "")
		// no user set

		handler := new(ReportProcessorHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("user with empty ID returns unauthorized error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		body := []byte(`{"filedata":"","filename":"x.csv","bankid":"b"}`)
		c, _ := testutil.NewGinContext("POST", "/api/v0/report/process", body, "")
		c.Set("user", &auth.UserClaims{ID: ""})

		handler := new(ReportProcessorHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("malformed JSON body returns bind error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		body := []byte(`{not-json`)
		c, _ := testutil.NewGinContext("POST", "/api/v0/report/process", body, "")
		testutil.SetAuthenticatedUser(c, "user-1", false)

		handler := new(ReportProcessorHandler).InitFromBuilder(mockBuilder)
		err := handler.Handler(c)

		require.Error(t, err)
	})
}

// --- processAsync ------------------------------------------------------------

func TestReportProcessorHandler_processAsync(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("calls CreateReportAsync with the right payload and returns queued", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockTask := taskmocks.NewMockTask(ctrl)

		mockBuilder.EXPECT().GetTask().Return(mockTask)

		expectedPayload := types.ReportProcessQueuePayload{
			UserId:   "user-1",
			BankId:   "bank-1",
			FilePath: "/tmp/upload_user-1_report.csv",
		}
		mockTask.EXPECT().CreateReportAsync(expectedPayload).Return(nil)

		c, w := testutil.NewGinContext("POST", "/api/v0/report/process", nil, "")

		handler := new(ReportProcessorHandler).InitFromBuilder(mockBuilder)
		req := Request{UserId: "user-1", BankId: "bank-1"}
		err := handler.processAsync(req, "/tmp/upload_user-1_report.csv", c)

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
		mockTask.EXPECT().CreateReportAsync(gomock.Any()).Return(fmt.Errorf("queue unavailable"))

		c, _ := testutil.NewGinContext("POST", "/api/v0/report/process", nil, "")

		handler := new(ReportProcessorHandler).InitFromBuilder(mockBuilder)
		req := Request{UserId: "user-1", BankId: "bank-1"}
		err := handler.processAsync(req, "/tmp/some-file.csv", c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "queue unavailable")
	})
}

// --- processSync (MockHTTPClient) --------------------------------------------

func writeReportTempFile(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "processsync_test_*.csv")
	require.NoError(t, err)
	_, err = f.WriteString("data")
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

func TestReportProcessorHandler_processSync(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("200 OK from DPW returns success and cleans up temp file", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockHTTPClient := httpclientmocks.NewMockHTTPClient(ctrl)

		mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{DPWHost: "dpw-host", DPWPort: "9000"}).AnyTimes()
		mockBuilder.EXPECT().GetHTTPClient().Return(mockHTTPClient).AnyTimes()

		tempFile := writeReportTempFile(t)

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

		c, w := testutil.NewGinContext("POST", "/api/v0/report/process", nil, "")

		handler := new(ReportProcessorHandler).InitFromBuilder(mockBuilder)
		req := Request{UserId: "user-1", BankId: "bank-1"}
		err := handler.processSync(req, tempFile, c)

		require.NoError(t, err)
		assert.Equal(t, 200, w.Code)

		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, "Report processed successfully", body["status"])

		assert.Equal(t, "POST", gotMethod)
		assert.Equal(t, "http://dpw-host:9000/dpw/v0/process_report", gotURL)
		assert.Equal(t, "application/json", gotContentType)

		var sentReq types.ReportProcessQueuePayload
		require.NoError(t, json.Unmarshal(gotBody, &sentReq))
		assert.Equal(t, "user-1", sentReq.UserId, "UserId must reach DPW -- it's tagged json:\"-\" on the local Request type (to stop a spoofed UserId in the inbound client body), so the outbound payload must NOT reuse that type or UserId silently vanishes")
		assert.Equal(t, "bank-1", sentReq.BankId)
		assert.Equal(t, tempFile, sentReq.FilePath)

		// processSync defers os.Remove(tempFilePath).
		_, statErr := os.Stat(tempFile)
		assert.True(t, os.IsNotExist(statErr), "temp file should have been cleaned up")
	})

	t.Run("non-200 status from DPW returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockHTTPClient := httpclientmocks.NewMockHTTPClient(ctrl)

		mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{DPWHost: "dpw-host", DPWPort: "9000"}).AnyTimes()
		mockBuilder.EXPECT().GetHTTPClient().Return(mockHTTPClient).AnyTimes()

		mockHTTPClient.EXPECT().Do(gomock.Any()).Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil)

		tempFile := writeReportTempFile(t)
		c, _ := testutil.NewGinContext("POST", "/api/v0/report/process", nil, "")

		handler := new(ReportProcessorHandler).InitFromBuilder(mockBuilder)
		req := Request{UserId: "user-1", BankId: "bank-1"}
		err := handler.processSync(req, tempFile, c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "500")
		assert.Contains(t, err.Error(), "DPW")

		// Even on error, the temp file cleanup defer must have run.
		_, statErr := os.Stat(tempFile)
		assert.True(t, os.IsNotExist(statErr))
	})

	t.Run("network error from HTTPClient is propagated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockHTTPClient := httpclientmocks.NewMockHTTPClient(ctrl)

		mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{DPWHost: "dpw-host", DPWPort: "9000"}).AnyTimes()
		mockBuilder.EXPECT().GetHTTPClient().Return(mockHTTPClient).AnyTimes()

		mockHTTPClient.EXPECT().Do(gomock.Any()).Return(nil, fmt.Errorf("connection refused"))

		tempFile := writeReportTempFile(t)
		c, _ := testutil.NewGinContext("POST", "/api/v0/report/process", nil, "")

		handler := new(ReportProcessorHandler).InitFromBuilder(mockBuilder)
		req := Request{UserId: "user-1", BankId: "bank-1"}
		err := handler.processSync(req, tempFile, c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "connection refused")
	})
}

// TestReportProcessorHandler_processSync_RealServer additionally exercises
// processSync against a real httptest server and the concrete
// httpclient.DefaultHTTPClient wrapper, to prove the injected HTTPClient
// interface is wired correctly end-to-end (not just against the mock).
func TestReportProcessorHandler_processSync_RealServer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/dpw/v0/process_report", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	serverURL := server.URL
	host := strings.TrimPrefix(serverURL, "http://")
	parts := strings.SplitN(host, ":", 2)
	require.Len(t, parts, 2)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockBuilder := buildermocks.NewMockBuilder(ctrl)
	mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{DPWHost: parts[0], DPWPort: parts[1]}).AnyTimes()
	mockBuilder.EXPECT().GetHTTPClient().Return(httpclient.New(server.Client())).AnyTimes()

	tempFile := writeReportTempFile(t)
	c, w := testutil.NewGinContext("POST", "/api/v0/report/process", nil, "")

	handler := new(ReportProcessorHandler).InitFromBuilder(mockBuilder)
	req := Request{UserId: "user-1", BankId: "bank-1"}
	err := handler.processSync(req, tempFile, c)

	require.NoError(t, err)
	assert.Equal(t, 200, w.Code)
}
