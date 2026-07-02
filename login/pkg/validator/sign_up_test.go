package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/verasthiago/verancial/shared/models"
)

func TestSignUpRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		user     *models.User
		wantErrs []string
	}{
		{
			name:     "valid name, email and password passes",
			user:     &models.User{Name: "Jane Doe", Email: "jane@example.com", Password: "secret"},
			wantErrs: nil,
		},
		{
			name:     "all fields missing",
			user:     &models.User{},
			wantErrs: []string{"The name field is required", "The email field is required", "The password field is required"},
		},
		{
			name:     "missing name only",
			user:     &models.User{Email: "jane@example.com", Password: "secret"},
			wantErrs: []string{"The name field is required"},
		},
		{
			name:     "missing email only",
			user:     &models.User{Name: "Jane Doe", Password: "secret"},
			wantErrs: []string{"The email field is required"},
		},
		{
			name:     "missing password only",
			user:     &models.User{Name: "Jane Doe", Email: "jane@example.com"},
			wantErrs: []string{"The password field is required"},
		},
		{
			name:     "invalid email format",
			user:     &models.User{Name: "Jane Doe", Email: "not-an-email", Password: "secret"},
			wantErrs: []string{"The email field must be a valid email address"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &SignUpRequest{User: tt.user}
			got := req.Validate()
			if tt.wantErrs == nil {
				assert.Empty(t, got)
			} else {
				assert.Equal(t, tt.wantErrs, got)
			}
		})
	}
}
