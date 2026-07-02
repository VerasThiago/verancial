// Package testutil provides small shared helpers for api unit tests: JWT
// generation and gin test-context construction. Kept deliberately tiny so
// each handler test package stays in control of its own mock wiring.
package testutil

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/verasthiago/verancial/shared/auth"
	"github.com/verasthiago/verancial/shared/models"
)

const TestJwtKey = "test-jwt-signing-key"

// GenerateToken builds a signed JWT for the given user id/admin flag,
// expiring at expiresAt.
func GenerateToken(t *testing.T, jwtKey string, userID string, isAdmin bool, expiresAt time.Time) string {
	t.Helper()
	user := &models.User{ID: userID, IsAdmin: isAdmin}
	token, err := auth.GenerateJWT(user, jwtKey, expiresAt)
	if err != nil {
		t.Fatalf("failed to generate test jwt: %v", err)
	}
	return token
}

// GenerateValidToken returns a JWT for userID valid for one hour, signed
// with TestJwtKey.
func GenerateValidToken(t *testing.T, userID string) string {
	t.Helper()
	return GenerateToken(t, TestJwtKey, userID, false, time.Now().Add(time.Hour))
}

// GenerateExpiredToken returns a JWT for userID that already expired.
func GenerateExpiredToken(t *testing.T, userID string) string {
	t.Helper()
	return GenerateToken(t, TestJwtKey, userID, false, time.Now().Add(-time.Hour))
}

// GenerateTokenSignedWithNone returns an "alg: none" token, used to test
// that the algorithm-confusion defense in shared/auth rejects it.
func GenerateTokenSignedWithNone(t *testing.T, userID string) string {
	t.Helper()
	claims := &auth.JWTClaim{
		User: &auth.UserClaims{ID: userID},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to generate none-alg test jwt: %v", err)
	}
	return signed
}

// NewGinContext builds a gin.Context/ResponseRecorder pair in test mode
// with an optional JSON body and Authorization header. Note: the app's
// convention is that the raw Authorization header IS the JWT (no "Bearer "
// prefix) -- shared/auth uses dgrijalva/jwt-go's Parse, which errors on a
// "bearer " prefix.
func NewGinContext(method, target string, body []byte, authHeader string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, target, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, target, nil)
	}

	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	c.Request = req
	return c, w
}

// SetAuthenticatedUser mimics what the auth middleware sets on the gin
// context after successful token validation.
func SetAuthenticatedUser(c *gin.Context, userID string, isAdmin bool) {
	c.Set("user", &auth.UserClaims{ID: userID, IsAdmin: isAdmin})
	c.Set("userId", userID)
}
