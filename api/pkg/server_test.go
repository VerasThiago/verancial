package pkg

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/verasthiago/verancial/api/pkg/builder"
	buildermocks "github.com/verasthiago/verancial/api/pkg/builder/mocks"
	sharedflags "github.com/verasthiago/verancial/shared/flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestServer_InitFromBuilder(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockBuilder := buildermocks.NewMockBuilder(ctrl)
	mockBuilder.EXPECT().GetFlags().Return(&builder.Flags{}).AnyTimes()
	mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{JwtKey: "secret"}).AnyTimes()

	s := new(Server).InitFromBuilder(mockBuilder)

	require.NotNil(t, s)
	assert.NotNil(t, s.ReportProcessorAPI)
	assert.NotNil(t, s.AppIntegrationAPI)
	assert.NotNil(t, s.DashboardAPI)
	assert.NotNil(t, s.BankAPI)
	assert.NotNil(t, s.ListTransactionsAPI)
	assert.NotNil(t, s.AuthAPI)
}

func TestServer_SetupRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)

	mockBuilder := buildermocks.NewMockBuilder(ctrl)
	mockBuilder.EXPECT().GetFlags().Return(&builder.Flags{}).AnyTimes()
	mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{JwtKey: "secret"}).AnyTimes()

	s := new(Server).InitFromBuilder(mockBuilder)
	router := s.SetupRouter()
	require.NotNil(t, router)

	// Every route below is registered behind the auth middleware, which
	// rejects a request with no Authorization header before it ever reaches
	// the underlying handler. A 401 (not 404) proves the route exists and
	// is wired to the auth middleware; a real success-path response is
	// covered separately by each handler's own tests.
	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v0/report/process"},
		{http.MethodPost, "/api/v0/app-integration/generate"},
		{http.MethodGet, "/api/v0/dashboard/user"},
		{http.MethodGet, "/api/v0/transaction/list/bank-1"},
		{http.MethodGet, "/api/v0/bank/bank-1"},
	}

	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestServer_Run_ReturnsListenError(t *testing.T) {
	// Run() blocks on a real listener when it succeeds, so the only part of
	// it safely exercisable from a unit test is the error path: an invalid
	// port makes net.Listen fail immediately and Run() returns that error
	// instead of blocking forever.
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)

	mockBuilder := buildermocks.NewMockBuilder(ctrl)
	mockBuilder.EXPECT().GetFlags().Return(&builder.Flags{Port: "not-a-port"}).AnyTimes()
	mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{JwtKey: "secret"}).AnyTimes()

	s := new(Server).InitFromBuilder(mockBuilder)

	assert.Error(t, s.Run())
}
