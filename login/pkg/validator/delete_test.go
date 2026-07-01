package validator

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDeleteRequest_Validate(t *testing.T) {
	validID := uuid.New().String()

	tests := []struct {
		name     string
		request  DeleteRequest
		wantErrs []string
	}{
		{
			name:     "valid uuid passes",
			request:  DeleteRequest{UserId: validID},
			wantErrs: nil,
		},
		{
			name:     "missing id fails required",
			request:  DeleteRequest{},
			wantErrs: []string{"The id field is required"},
		},
		{
			name:     "non-uuid id fails uuid rule",
			request:  DeleteRequest{UserId: "not-a-uuid"},
			wantErrs: []string{"The id field must contain valid UUID"},
		},
		{
			name:     "numeric-looking id still fails uuid rule",
			request:  DeleteRequest{UserId: "12345"},
			wantErrs: []string{"The id field must contain valid UUID"},
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
