package pkg

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/verasthiago/verancial/data-process-worker/pkg/builder"
	buildermocks "github.com/verasthiago/verancial/data-process-worker/pkg/builder/mocks"
)

func TestServer_InitFromBuilder(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockBuilder := buildermocks.NewMockBuilder(ctrl)

	s := new(Server).InitFromBuilder(mockBuilder)

	require.NotNil(t, s)
	assert.NotNil(t, s.ReportCreateAPI)
	assert.Equal(t, builder.Builder(mockBuilder), s.Builder)
}

func TestServer_SetupRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	mockBuilder := buildermocks.NewMockBuilder(ctrl)

	s := new(Server).InitFromBuilder(mockBuilder)
	router := s.SetupRouter()
	require.NotNil(t, router)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/dpw/v0/process_report", nil)
	router.ServeHTTP(w, req)

	// Malformed/empty body -> handler returns a bind error, not a 404 --
	// proves the route is registered and reachable (no auth on this
	// service-to-service endpoint).
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

func TestServer_RunSync_ReturnsListenError(t *testing.T) {
	// RunSync blocks on a real listener when it succeeds, so the only part
	// safely exercisable from a unit test is the error path: an invalid
	// port makes net.Listen fail immediately and Run() returns that error
	// instead of blocking forever.
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	mockBuilder := buildermocks.NewMockBuilder(ctrl)
	mockBuilder.EXPECT().GetFlags().Return(&builder.Flags{Port: "not-a-port", AsyncProcessing: false}).AnyTimes()

	s := new(Server).InitFromBuilder(mockBuilder)

	assert.Error(t, s.Run())
}
