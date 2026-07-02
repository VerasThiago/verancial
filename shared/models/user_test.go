package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestUser_HashPassword(t *testing.T) {
	t.Run("hashes password and stores a bcrypt hash different from the plaintext", func(t *testing.T) {
		u := &User{}
		err := u.HashPassword("s3cr3t-password")

		require.NoError(t, err)
		assert.NotEmpty(t, u.Password)
		assert.NotEqual(t, "s3cr3t-password", u.Password)

		// Confirm it's actually a valid bcrypt hash of the input.
		assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("s3cr3t-password")))
	})

	t.Run("password exceeding bcrypt's 72 byte limit is silently truncated by the pinned x/crypto version", func(t *testing.T) {
		// golang.org/x/crypto/bcrypt only started rejecting >72-byte inputs
		// with ErrPasswordTooLong in a later release than the one pinned in
		// shared/go.mod; on this version GenerateFromPassword truncates
		// instead of erroring. This test documents current behavior rather
		// than asserting a stricter contract the pinned dependency doesn't
		// provide.
		u := &User{}
		longPassword := make([]byte, 100)
		for i := range longPassword {
			longPassword[i] = 'a'
		}

		err := u.HashPassword(string(longPassword))

		require.NoError(t, err)
		assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(u.Password), longPassword))
	})
}

func TestUser_CheckPassword(t *testing.T) {
	t.Run("correct password succeeds", func(t *testing.T) {
		u := &User{}
		require.NoError(t, u.HashPassword("correct-horse-battery-staple"))

		err := u.CheckPassword("correct-horse-battery-staple")

		assert.NoError(t, err)
	})

	t.Run("incorrect password fails", func(t *testing.T) {
		u := &User{}
		require.NoError(t, u.HashPassword("correct-horse-battery-staple"))

		err := u.CheckPassword("wrong-password")

		assert.Error(t, err)
	})

	t.Run("malformed stored hash fails", func(t *testing.T) {
		u := &User{Password: "not-a-real-bcrypt-hash"}

		err := u.CheckPassword("anything")

		assert.Error(t, err)
	})

	t.Run("empty stored hash fails", func(t *testing.T) {
		u := &User{Password: ""}

		err := u.CheckPassword("anything")

		assert.Error(t, err)
	})
}

func TestUser_BeforeCreate(t *testing.T) {
	t.Run("assigns a UUID, lowercases email, title-cases name, and marks verified", func(t *testing.T) {
		u := &User{
			Email: "JANE.DOE@EXAMPLE.COM",
			Name:  "jane doe",
		}

		err := u.BeforeCreate(nil)

		require.NoError(t, err)
		assert.NotEmpty(t, u.ID)
		assert.Equal(t, "jane.doe@example.com", u.Email)
		assert.Equal(t, "Jane Doe", u.Name)
		assert.True(t, u.IsVerified)
	})

	t.Run("generates a different ID on each call", func(t *testing.T) {
		u1 := &User{}
		u2 := &User{}
		require.NoError(t, u1.BeforeCreate(nil))
		require.NoError(t, u2.BeforeCreate(nil))

		assert.NotEqual(t, u1.ID, u2.ID)
	})
}
