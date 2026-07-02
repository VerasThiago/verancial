package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	buildermocks "github.com/verasthiago/verancial/api/pkg/builder/mocks"
	"github.com/verasthiago/verancial/api/pkg/middlewares"
	"github.com/verasthiago/verancial/api/pkg/testutil"
	sharedflags "github.com/verasthiago/verancial/shared/flags"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelloPrivateHandler_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilder := buildermocks.NewMockBuilder(ctrl)

	c, w := testutil.NewGinContext("GET", "/api/v0/hello-private", nil, "")

	handler := new(HelloPrivateHandler).InitFromBuilder(mockBuilder)
	handler.Handler(c)

	assert.Equal(t, 200, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "success", body["status"])
	assert.Equal(t, "hello private!", body["message"])
}

// TestHelloPrivateHandler_RequiresAuth wires HelloPrivateHandler behind the
// real AuthUserHandler middleware in a standalone router (mirroring how a
// private route would be registered) to prove that, when protected by the
// auth middleware, unauthenticated requests never reach the handler.
func TestHelloPrivateHandler_RequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtKey := "hello-private-key"
	authHandler := new(middlewares.AuthUserHandler).InitFromFlags(nil, &sharedflags.SharedFlags{JwtKey: jwtKey})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockBuilder := buildermocks.NewMockBuilder(ctrl)
	helloPrivate := new(HelloPrivateHandler).InitFromBuilder(mockBuilder)

	router := gin.New()
	router.GET("/private/hello", authHandler.Handler(), func(c *gin.Context) {
		helloPrivate.Handler(c)
	})

	t.Run("without token is rejected before reaching handler", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/private/hello", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("with valid token reaches handler", func(t *testing.T) {
		token := testutil.GenerateToken(t, jwtKey, "user-1", false, time.Now().Add(time.Hour))

		req := httptest.NewRequest("GET", "/private/hello", nil)
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, "hello private!", body["message"])
	})
}
