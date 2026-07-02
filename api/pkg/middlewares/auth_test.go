package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/verasthiago/verancial/api/pkg/testutil"
	"github.com/verasthiago/verancial/shared/auth"
	sharedflags "github.com/verasthiago/verancial/shared/flags"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newAuthTestRouter wires the real AuthUserHandler middleware in front of a
// handler that echoes back what was set on the gin context, so we can
// assert both the HTTP-level outcome (status/abort) and that "user"/"userId"
// were populated correctly for downstream handlers.
func newAuthTestRouter(jwtKey string) *gin.Engine {
	gin.SetMode(gin.TestMode)

	authHandler := new(AuthUserHandler).InitFromFlags(nil, &sharedflags.SharedFlags{JwtKey: jwtKey})

	router := gin.New()
	router.GET("/protected", authHandler.Handler(), func(c *gin.Context) {
		userObj, _ := c.Get("user")
		userIdObj, _ := c.Get("userId")

		user, _ := userObj.(*auth.UserClaims)
		c.JSON(http.StatusOK, gin.H{
			"user":   user,
			"userId": userIdObj,
		})
	})
	return router
}

func TestAuthUserHandler_Handler(t *testing.T) {
	jwtKey := "auth-middleware-test-key"
	router := newAuthTestRouter(jwtKey)

	t.Run("valid token proceeds with user set in context", func(t *testing.T) {
		userID := "user-42"
		token := testutil.GenerateToken(t, jwtKey, userID, true, time.Now().Add(time.Hour))

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, userID, body["userId"])
		user := body["user"].(map[string]interface{})
		assert.Equal(t, userID, user["id"])
		assert.Equal(t, true, user["isadmin"])
	})

	t.Run("missing Authorization header returns 401 and aborts", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Contains(t, body["error"], "access token")
	})

	t.Run("malformed token returns 401 and aborts", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "this-is-not-a-jwt")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("expired token returns 401 and aborts", func(t *testing.T) {
		token := testutil.GenerateToken(t, jwtKey, "user-1", false, time.Now().Add(-time.Hour))

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Contains(t, body["error"], "expired")
	})

	t.Run("token signed with wrong key returns 401 and aborts", func(t *testing.T) {
		token := testutil.GenerateToken(t, "a-completely-different-key", "user-1", false, time.Now().Add(time.Hour))

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("none-alg token is rejected", func(t *testing.T) {
		token := testutil.GenerateTokenSignedWithNone(t, "user-1")

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Bearer-prefixed token is rejected (app convention is raw JWT)", func(t *testing.T) {
		token := testutil.GenerateToken(t, jwtKey, "user-1", false, time.Now().Add(time.Hour))

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verified empirically: dgrijalva/jwt-go's parser explicitly rejects
		// token strings containing a "bearer " prefix, and this middleware
		// passes the raw header straight through without stripping it.
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
