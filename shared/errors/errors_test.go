package errors

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenericError_Error(t *testing.T) {
	t.Run("formats code, original error, metadata and message", func(t *testing.T) {
		origErr := errors.New("boom")
		e := GenericError{
			Code:     STATUS_BAD_REQUEST,
			Type:     INVALID_INPUT.Type,
			Err:      origErr,
			Message:  "bad request",
			MetaData: map[string]interface{}{"field": "email"},
		}

		s := e.Error()

		assert.Contains(t, s, "boom")
		assert.Contains(t, s, "bad request")
		assert.Contains(t, s, "field")
	})

	t.Run("handles nil error and metadata gracefully", func(t *testing.T) {
		e := GenericError{
			Code:    STATUS_INTERNAL_SERVER_ERROR,
			Message: "oops",
		}

		s := e.Error()
		assert.Contains(t, s, "oops")
	})
}

func TestGenericError_GenerateJsonResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("writes expected status code and JSON body", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		e := GenericError{
			Code:    STATUS_NOT_FOUND,
			Type:    DATA_NOT_FOUND.Type,
			Message: DATA_NOT_FOUND.Message,
			MetaData: map[string]interface{}{
				"variables": []map[string]interface{}{{"path": "models.user"}},
			},
		}

		e.GenerateJsonResponse(c)

		assert.Equal(t, 404, w.Code)
		body := w.Body.String()
		assert.Contains(t, body, `"status":"failed"`)
		assert.Contains(t, body, DATA_NOT_FOUND.Type)
		assert.Contains(t, body, DATA_NOT_FOUND.Message)
		assert.Contains(t, body, "models.user")
	})

	t.Run("aborts the gin context", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		e := GenericError{Code: STATUS_INTERNAL_SERVER_ERROR, Type: GENERIC_ERROR.Type, Message: GENERIC_ERROR.Message}
		e.GenerateJsonResponse(c)

		assert.True(t, c.IsAborted())
	})
}

func TestGenericError_ImplementsJsonError(t *testing.T) {
	var _ JsonError = GenericError{}
}

func TestGenericError_AsError(t *testing.T) {
	// GenericError must satisfy the error interface so it can be returned
	// from functions returning `error` (e.g. HandleDuplicateError).
	var err error = GenericError{Code: STATUS_BAD_REQUEST, Message: "x"}
	require.Error(t, err)
}
