package errors

import (
	stderrors "errors"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/verasthiago/verancial/shared/models"
	"gorm.io/gorm"
)

func TestErrorRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("no error from route: nothing is written", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		route := func(c *gin.Context) error { return nil }
		handler := ErrorRoute(route)
		handler(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, 200, w.Code) // default recorder code when nothing written
	})

	t.Run("route returning a plain error triggers generic 500 response", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		var route models.Route = func(c *gin.Context) error { return stderrors.New("db exploded") }
		handler := ErrorRoute(route)
		handler(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, 500, w.Code)
		assert.Contains(t, w.Body.String(), GENERIC_ERROR.Type)
	})

	t.Run("route returning a GenericError uses its own status/body", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		var route models.Route = func(c *gin.Context) error {
			return GenericError{Code: STATUS_BAD_REQUEST, Type: INVALID_INPUT.Type, Message: "bad"}
		}
		handler := ErrorRoute(route)
		handler(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, 400, w.Code)
		assert.Contains(t, w.Body.String(), INVALID_INPUT.Type)
	})
}

func TestRecoveryHandler_StringError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	recoveryHandler(c, "raw string failure")

	assert.True(t, c.IsAborted())
	assert.Equal(t, 500, w.Code)
	assert.Contains(t, w.Body.String(), "raw string failure")
}

func TestRecoveryHandler_UnknownType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Not an error, JsonError, or string -> no branch matches, nothing written.
	recoveryHandler(c, 12345)

	assert.False(t, c.IsAborted())
}

func TestCreateGenericErrorFromValidateError(t *testing.T) {
	t.Run("joins multiple validation errors with newline", func(t *testing.T) {
		e := CreateGenericErrorFromValidateError([]string{"email is required", "password too short"})

		assert.Equal(t, StatusCode(STATUS_BAD_REQUEST), e.Code)
		assert.Equal(t, INVALID_INPUT.Type, e.Type)
		assert.Equal(t, "email is required\npassword too short", e.Message)
	})

	t.Run("single error message", func(t *testing.T) {
		e := CreateGenericErrorFromValidateError([]string{"only one issue"})
		assert.Equal(t, "only one issue", e.Message)
	})

	t.Run("empty slice produces empty message", func(t *testing.T) {
		e := CreateGenericErrorFromValidateError([]string{})
		assert.Equal(t, "", e.Message)
		assert.Equal(t, StatusCode(STATUS_BAD_REQUEST), e.Code)
	})
}

func TestIsNotFoundError(t *testing.T) {
	assert.True(t, IsNotFoundError(gorm.ErrRecordNotFound))
	assert.False(t, IsNotFoundError(stderrors.New("some other error")))
	assert.False(t, IsNotFoundError(nil))
}

func TestIsDuplicatedKeyError(t *testing.T) {
	t.Run("pg duplicate key error code detected", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: PSQL_DUPLICATED_KEY_ERROR_CODE, Message: `duplicate key value violates unique constraint "users_email_key"`}
		assert.True(t, IsDuplicatedKeyError(pgErr))
	})

	t.Run("pg error with a different code is not a duplicate", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: "23503", Message: "foreign key violation"}
		assert.False(t, IsDuplicatedKeyError(pgErr))
	})

	t.Run("non pg error returns false", func(t *testing.T) {
		assert.False(t, IsDuplicatedKeyError(stderrors.New("generic error")))
	})

	t.Run("nil error returns false", func(t *testing.T) {
		assert.False(t, IsDuplicatedKeyError(nil))
	})

	t.Run("wrapped pg error is still detected via errors.As", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: PSQL_DUPLICATED_KEY_ERROR_CODE}
		wrapped := fmt.Errorf("wrapped: %w", pgErr)
		assert.True(t, IsDuplicatedKeyError(wrapped))
	})
}

func TestBuildI18NPath(t *testing.T) {
	assert.Equal(t, "models.user", BuildI18NPath(I18N_MODELS, "user"))
	assert.Equal(t, "fields.email", BuildI18NPath(I18N_FIELDS, "email"))
	assert.Equal(t, ".", BuildI18NPath("", ""))
}

func TestHandleDuplicateError(t *testing.T) {
	t.Run("nil error passes through", func(t *testing.T) {
		require.NoError(t, HandleDuplicateError(nil))
	})

	t.Run("non-duplicate error passes through unchanged", func(t *testing.T) {
		orig := stderrors.New("some other failure")
		err := HandleDuplicateError(orig)
		assert.Equal(t, orig, err)
	})

	t.Run("duplicate key error on mapped column produces friendly GenericError", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Code:    PSQL_DUPLICATED_KEY_ERROR_CODE,
			Message: `duplicate key value violates unique constraint "users_email_key"`,
		}

		err := HandleDuplicateError(pgErr)

		require.Error(t, err)
		genericErr, ok := err.(GenericError)
		require.True(t, ok)
		assert.Equal(t, StatusCode(STATUS_BAD_REQUEST), genericErr.Code)
		assert.Equal(t, DATA_ALREADY_BEGIN_USED.Type, genericErr.Type)

		meta, ok := genericErr.MetaData.(map[string]interface{})
		require.True(t, ok)
		variables, ok := meta["variables"].([]map[string]interface{})
		require.True(t, ok)
		require.Len(t, variables, 1)
		assert.Equal(t, "fields.email", variables[0]["path"])
	})

	t.Run("duplicate key error on unmapped column falls back to UNDEFINED", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Code:    PSQL_DUPLICATED_KEY_ERROR_CODE,
			Message: `duplicate key value violates unique constraint "some_other_key"`,
		}

		err := HandleDuplicateError(pgErr)

		genericErr, ok := err.(GenericError)
		require.True(t, ok)
		meta := genericErr.MetaData.(map[string]interface{})
		variables := meta["variables"].([]map[string]interface{})
		assert.Equal(t, "fields.UNDEFINED", variables[0]["path"])
	})
}

func TestHandleDataNotFoundError(t *testing.T) {
	t.Run("nil error passes through", func(t *testing.T) {
		require.NoError(t, HandleDataNotFoundError(nil, "user"))
	})

	t.Run("non-not-found error passes through unchanged", func(t *testing.T) {
		orig := stderrors.New("connection refused")
		err := HandleDataNotFoundError(orig, "user")
		assert.Equal(t, orig, err)
	})

	t.Run("gorm record not found produces friendly GenericError", func(t *testing.T) {
		err := HandleDataNotFoundError(gorm.ErrRecordNotFound, "user")

		require.Error(t, err)
		genericErr, ok := err.(GenericError)
		require.True(t, ok)
		assert.Equal(t, StatusCode(STATUS_NOT_FOUND), genericErr.Code)
		assert.Equal(t, DATA_NOT_FOUND.Type, genericErr.Type)

		meta := genericErr.MetaData.(map[string]interface{})
		variables := meta["variables"].([]map[string]interface{})
		assert.Equal(t, "models.user", variables[0]["path"])
	})
}
