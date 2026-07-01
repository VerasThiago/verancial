package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignInRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		request  SignInRequest
		wantErrs []string
	}{
		{
			name:     "valid email and password passes",
			request:  SignInRequest{Email: "jane@example.com", Password: "secret"},
			wantErrs: nil,
		},
		{
			name:     "missing email and password",
			request:  SignInRequest{},
			wantErrs: []string{"The email field is required", "The password field is required"},
		},
		{
			name:     "missing email only",
			request:  SignInRequest{Password: "secret"},
			wantErrs: []string{"The email field is required"},
		},
		{
			name:     "missing password only",
			request:  SignInRequest{Email: "jane@example.com"},
			wantErrs: []string{"The password field is required"},
		},
		{
			name:     "invalid email format",
			request:  SignInRequest{Email: "not-an-email", Password: "secret"},
			wantErrs: []string{"The email field must be a valid email address"},
		},
		{
			name:     "email without domain",
			request:  SignInRequest{Email: "jane@", Password: "secret"},
			wantErrs: []string{"The email field must be a valid email address"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.request.Validate()
			if tt.wantErrs == nil {
				assert.Empty(t, got)
			} else {
				assert.Equal(t, tt.wantErrs, got)
			}
		})
	}
}
