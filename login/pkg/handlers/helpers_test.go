package handlers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sharederrors "github.com/verasthiago/verancial/shared/errors"
)

// assertGenericError asserts that err is a shared/errors.GenericError with
// the given HTTP status code and error Type. If wantMessage is non-empty it
// also asserts the exact Message; otherwise the message is left unchecked.
func assertGenericError(t *testing.T, err error, wantCode int, wantType, wantMessage string) {
	t.Helper()

	require.Error(t, err)
	genericErr, ok := err.(sharederrors.GenericError)
	require.Truef(t, ok, "expected error of type sharederrors.GenericError, got %T: %v", err, err)

	assert.EqualValues(t, wantCode, genericErr.Code)
	assert.Equal(t, wantType, genericErr.Type)
	if wantMessage != "" {
		assert.Equal(t, wantMessage, genericErr.Message)
	}
}

func errRecordNotFound() error {
	return errors.New("record not found")
}
