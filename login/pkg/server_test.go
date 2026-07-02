package pkg

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/verasthiago/verancial/login/pkg/builder"
	buildermocks "github.com/verasthiago/verancial/login/pkg/builder/mocks"
	sharedflags "github.com/verasthiago/verancial/shared/flags"
)

func TestServer_InitFromBuilder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilder := buildermocks.NewMockBuilder(ctrl)
	mockBuilder.EXPECT().GetFlags().Return(&builder.Flags{Port: "8080"}).AnyTimes()
	mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{JwtKey: "key"}).AnyTimes()

	server := new(Server).InitFromBuilder(mockBuilder)

	require.NotNil(t, server)
	assert.NotNil(t, server.LoginAPI)
	assert.NotNil(t, server.CreateAPI)
	assert.NotNil(t, server.DeleteAPI)
	assert.NotNil(t, server.UpdateAPI)
	assert.NotNil(t, server.AdminAPI)
	assert.Equal(t, builder.Builder(mockBuilder), server.Builder)
}

func TestServer_SetupRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilder := buildermocks.NewMockBuilder(ctrl)
	mockBuilder.EXPECT().GetFlags().Return(&builder.Flags{Port: "8080"}).AnyTimes()
	mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{JwtKey: "key"}).AnyTimes()

	server := new(Server).InitFromBuilder(mockBuilder)
	router := server.SetupRouter()
	require.NotNil(t, router)

	t.Run("public routes reach the handler (no auth required)", func(t *testing.T) {
		for _, path := range []string{"/login/v0/user/signin", "/login/v0/user/signup"} {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, path, nil)
			router.ServeHTTP(w, req)

			// Malformed/empty body -> handler returns a bind error, not a
			// 404 -- proves the route is registered and reachable without auth.
			assert.NotEqual(t, http.StatusNotFound, w.Code, path)
			assert.NotEqual(t, http.StatusUnauthorized, w.Code, path)
		}
	})

	t.Run("admin routes require auth", func(t *testing.T) {
		cases := []struct {
			method string
			path   string
		}{
			{http.MethodDelete, "/login/v0/admin/delete"},
			{http.MethodPut, "/login/v0/admin/update"},
		}
		for _, tc := range cases {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code, tc.path)
		}
	})
}

func TestServer_Run_ReturnsListenError(t *testing.T) {
	// Run() blocks on a real listener when it succeeds, so the only part of
	// it safely exercisable from a unit test is the error path: an invalid
	// port makes net.Listen fail immediately and Run() returns that error
	// instead of blocking forever.
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilder := buildermocks.NewMockBuilder(ctrl)
	mockBuilder.EXPECT().GetFlags().Return(&builder.Flags{Port: "not-a-port"}).AnyTimes()
	mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{JwtKey: "key"}).AnyTimes()

	server := new(Server).InitFromBuilder(mockBuilder)

	assert.Error(t, server.Run())
}
