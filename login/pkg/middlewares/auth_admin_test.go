package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/verasthiago/verancial/login/pkg/builder"
	"github.com/verasthiago/verancial/shared/auth"
	sharedflags "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/shared/models"
)

func init() {
	gin.SetMode(gin.TestMode)
}

const testJwtKey = "test-jwt-key"

func newAuthHandler() *AuthUserHandler {
	return new(AuthUserHandler).InitFromFlags(&builder.Flags{Port: "8080"}, &sharedflags.SharedFlags{JwtKey: testJwtKey})
}

// buildRouter wires the auth middleware in front of a trivial handler that
// records whether it was reached, so tests can assert both the HTTP response
// and whether context.Next() was actually called.
func buildRouter(handler gin.HandlerFunc) (*gin.Engine, *bool) {
	reached := false
	router := gin.New()
	router.Use(handler)
	router.GET("/protected", func(c *gin.Context) {
		reached = true
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return router, &reached
}

func TestAuthUserHandler_Handler(t *testing.T) {
	t.Run("valid admin JWT proceeds to the next handler", func(t *testing.T) {
		authHandler := newAuthHandler()
		router, reached := buildRouter(authHandler.Handler())

		token, err := auth.GenerateJWT(&models.User{ID: "admin-1", IsAdmin: true}, testJwtKey, time.Now().Add(1*time.Hour))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, *reached)
	})

	t.Run("missing Authorization header is rejected with 401", func(t *testing.T) {
		authHandler := newAuthHandler()
		router, reached := buildRouter(authHandler.Handler())

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, *reached)
		assert.Contains(t, w.Body.String(), "request does not contain an access token")
	})

	t.Run("invalid/garbage token is rejected with 401", func(t *testing.T) {
		authHandler := newAuthHandler()
		router, reached := buildRouter(authHandler.Handler())

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "not-a-real-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, *reached)
	})

	t.Run("expired token is rejected with 401", func(t *testing.T) {
		authHandler := newAuthHandler()
		router, reached := buildRouter(authHandler.Handler())

		token, err := auth.GenerateJWT(&models.User{ID: "admin-1", IsAdmin: true}, testJwtKey, time.Now().Add(-1*time.Hour))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, *reached)
		assert.Contains(t, w.Body.String(), "token is expired")
	})

	t.Run("token signed with the wrong key is rejected with 401", func(t *testing.T) {
		authHandler := newAuthHandler()
		router, reached := buildRouter(authHandler.Handler())

		token, err := auth.GenerateJWT(&models.User{ID: "admin-1", IsAdmin: true}, "a-different-key", time.Now().Add(1*time.Hour))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, *reached)
	})

	t.Run("non-admin token is rejected with 401 and admin-specific message", func(t *testing.T) {
		authHandler := newAuthHandler()
		router, reached := buildRouter(authHandler.Handler())

		token, err := auth.GenerateJWT(&models.User{ID: "user-1", IsAdmin: false}, testJwtKey, time.Now().Add(1*time.Hour))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, *reached)
		assert.Contains(t, w.Body.String(), "user isn't admin")
	})
}

func TestAuthUserHandler_InitFromFlags(t *testing.T) {
	flags := &builder.Flags{Port: "9090"}
	sharedFlags := &sharedflags.SharedFlags{JwtKey: "k"}

	handler := new(AuthUserHandler).InitFromFlags(flags, sharedFlags)

	require.NotNil(t, handler)
	assert.Same(t, flags, handler.Flags)
	assert.Same(t, sharedFlags, handler.SharedFlags)
}
