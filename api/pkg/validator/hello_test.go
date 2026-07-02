package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignInRequest_Validate(t *testing.T) {
	t.Run("valid email and password produce no validation errors", func(t *testing.T) {
		req := &SignInRequest{Email: "user@example.com", Password: "hunter2"}
		errs := req.Validate()
		assert.Empty(t, errs)
	})

	t.Run("missing email and password produce errors for both", func(t *testing.T) {
		req := &SignInRequest{}
		errs := req.Validate()
		require := assert.New(t)
		require.NotEmpty(errs)
		require.Len(errs, 2)
	})

	t.Run("missing email only produces one error", func(t *testing.T) {
		req := &SignInRequest{Password: "hunter2"}
		errs := req.Validate()
		assert.Len(t, errs, 1)
	})

	t.Run("missing password only produces one error", func(t *testing.T) {
		req := &SignInRequest{Email: "user@example.com"}
		errs := req.Validate()
		assert.Len(t, errs, 1)
	})

	t.Run("malformed email produces a validation error", func(t *testing.T) {
		req := &SignInRequest{Email: "not-an-email", Password: "hunter2"}
		errs := req.Validate()
		assert.NotEmpty(t, errs)
	})

	t.Run("empty password string is treated as missing", func(t *testing.T) {
		req := &SignInRequest{Email: "user@example.com", Password: ""}
		errs := req.Validate()
		assert.NotEmpty(t, errs)
	})
}
