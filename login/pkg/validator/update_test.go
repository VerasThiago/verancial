package validator

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/verasthiago/verancial/shared/models"
)

func TestUpdateRequest_Validate(t *testing.T) {
	validID := uuid.New().String()

	tests := []struct {
		name     string
		user     *models.User
		wantErrs []string
	}{
		{
			name:     "valid id only (name/email/password optional) passes",
			user:     &models.User{ID: validID},
			wantErrs: nil,
		},
		{
			name:     "valid id, alpha name, valid email passes",
			user:     &models.User{ID: validID, Name: "Janet", Email: "jane@example.com"},
			wantErrs: nil,
		},
		{
			name:     "missing id fails required",
			user:     &models.User{},
			wantErrs: []string{"The id field is required"},
		},
		{
			name:     "non-uuid id fails uuid rule",
			user:     &models.User{ID: "not-a-uuid"},
			wantErrs: []string{"The id field must contain valid UUID"},
		},
		{
			name:     "name containing digits fails alpha rule",
			user:     &models.User{ID: validID, Name: "Janet123"},
			wantErrs: []string{"The name may only contain letters"},
		},
		{
			name:     "name containing a space fails alpha rule",
			user:     &models.User{ID: validID, Name: "Janet Doe"},
			wantErrs: []string{"The name may only contain letters"},
		},
		{
			name:     "invalid email format fails email rule",
			user:     &models.User{ID: validID, Email: "not-an-email"},
			wantErrs: []string{"The email field must be a valid email address"},
		},
		{
			name:     "multiple invalid fields report multiple errors",
			user:     &models.User{ID: "not-a-uuid", Name: "Janet123", Email: "bad"},
			wantErrs: []string{"The id field must contain valid UUID", "The name may only contain letters", "The email field must be a valid email address"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &UpdateRequest{User: tt.user}
			got := req.Validate()
			if tt.wantErrs == nil {
				assert.Empty(t, got)
			} else {
				assert.Equal(t, tt.wantErrs, got)
			}
		})
	}
}

func TestUpdateRequest_Validate_NilUser(t *testing.T) {
	// A nil embedded *models.User must not panic; govalidator treats the
	// missing fields as zero values, so "id" is reported as required.
	req := &UpdateRequest{}

	assert.NotPanics(t, func() {
		got := req.Validate()
		assert.Equal(t, []string{"The id field is required"}, got)
	})
}
