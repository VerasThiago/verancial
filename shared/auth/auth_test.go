package auth

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/verasthiago/verancial/shared/models"
)

func testUser() *models.User {
	return &models.User{
		ID:      "user-123",
		Name:    "Jane Doe",
		Email:   "jane@example.com",
		IsAdmin: true,
	}
}

func TestUserClaimsFromUser(t *testing.T) {
	t.Run("nil user returns nil claims", func(t *testing.T) {
		assert.Nil(t, userClaimsFromUser(nil))
	})

	t.Run("populated user maps only safe fields", func(t *testing.T) {
		user := testUser()
		user.Password = "super-secret-hash"

		claims := userClaimsFromUser(user)

		require.NotNil(t, claims)
		assert.Equal(t, user.ID, claims.ID)
		assert.Equal(t, user.IsAdmin, claims.IsAdmin)
	})
}

func TestGenerateJWT(t *testing.T) {
	t.Run("produces a parseable, validly signed token", func(t *testing.T) {
		user := testUser()
		expiration := time.Now().Add(time.Hour)

		tokenString, err := GenerateJWT(user, "my-signing-key", expiration)

		require.NoError(t, err)
		assert.NotEmpty(t, tokenString)

		claims, err := GetJWTClaimFromToken(tokenString, "my-signing-key")
		require.NoError(t, err)
		assert.Equal(t, user.ID, claims.User.ID)
		assert.Equal(t, user.IsAdmin, claims.User.IsAdmin)
		assert.Equal(t, expiration.Unix(), claims.ExpiresAt)
	})

	t.Run("nil user still produces a signed token with nil user claims", func(t *testing.T) {
		tokenString, err := GenerateJWT(nil, "my-signing-key", time.Now().Add(time.Hour))

		require.NoError(t, err)
		assert.NotEmpty(t, tokenString)
	})
}

func TestValidateToken(t *testing.T) {
	t.Run("valid, non-expired token passes", func(t *testing.T) {
		user := testUser()
		tokenString, err := GenerateJWT(user, "signing-key", time.Now().Add(time.Hour))
		require.NoError(t, err)

		err = ValidateToken(tokenString, "signing-key")
		assert.NoError(t, err)
	})

	t.Run("expired token is rejected", func(t *testing.T) {
		user := testUser()
		tokenString, err := GenerateJWT(user, "signing-key", time.Now().Add(-time.Hour))
		require.NoError(t, err)

		err = ValidateToken(tokenString, "signing-key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("wrong signing key is rejected", func(t *testing.T) {
		user := testUser()
		tokenString, err := GenerateJWT(user, "correct-key", time.Now().Add(time.Hour))
		require.NoError(t, err)

		err = ValidateToken(tokenString, "wrong-key")
		assert.Error(t, err)
	})

	t.Run("malformed token is rejected", func(t *testing.T) {
		err := ValidateToken("not-a-valid-jwt", "signing-key")
		assert.Error(t, err)
	})

	t.Run("tampered token is rejected", func(t *testing.T) {
		user := testUser()
		tokenString, err := GenerateJWT(user, "signing-key", time.Now().Add(time.Hour))
		require.NoError(t, err)

		// Flip the last character of the signature to corrupt it.
		tampered := tokenString[:len(tokenString)-1]
		if tokenString[len(tokenString)-1] == 'A' {
			tampered += "B"
		} else {
			tampered += "A"
		}

		err = ValidateToken(tampered, "signing-key")
		assert.Error(t, err)
	})

	t.Run("empty token string is rejected", func(t *testing.T) {
		err := ValidateToken("", "signing-key")
		assert.Error(t, err)
	})
}

func TestGetJWTClaimFromToken(t *testing.T) {
	t.Run("round-trips claims correctly", func(t *testing.T) {
		user := testUser()
		expiration := time.Now().Add(30 * time.Minute)
		tokenString, err := GenerateJWT(user, "key", expiration)
		require.NoError(t, err)

		claims, err := GetJWTClaimFromToken(tokenString, "key")

		require.NoError(t, err)
		require.NotNil(t, claims.User)
		assert.Equal(t, user.ID, claims.User.ID)
		assert.Equal(t, user.IsAdmin, claims.User.IsAdmin)
		assert.Equal(t, expiration.Unix(), claims.ExpiresAt)
	})

	t.Run("rejects tokens signed with the none algorithm", func(t *testing.T) {
		claims := &JWTClaim{
			User: &UserClaims{ID: "attacker", IsAdmin: true},
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
		tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		require.NoError(t, err)

		_, err = GetJWTClaimFromToken(tokenString, "any-key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected signing method")
	})

	t.Run("rejects a token whose claims carry no user", func(t *testing.T) {
		claims := &JWTClaim{
			User: nil,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("key"))
		require.NoError(t, err)

		_, err = GetJWTClaimFromToken(tokenString, "key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing user claims")
	})

	t.Run("rejects garbage input", func(t *testing.T) {
		_, err := GetJWTClaimFromToken("garbage.token.value", "key")
		assert.Error(t, err)
	})

	t.Run("rejects a token whose claims are wrong shape (map claims)", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"foo": "bar",
			"exp": time.Now().Add(time.Hour).Unix(),
		})
		tokenString, err := token.SignedString([]byte("key"))
		require.NoError(t, err)

		// This will actually fail to parse into JWTClaim entirely since
		// ParseWithClaims uses the concrete type passed in, so it should
		// still produce a JWTClaim with a nil User -> "missing user claims".
		_, err = GetJWTClaimFromToken(tokenString, "key")
		assert.Error(t, err)
	})
}
