package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/verasthiago/verancial/login/pkg/builder"
	buildermocks "github.com/verasthiago/verancial/login/pkg/builder/mocks"
	sharedflags "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/shared/models"
	repositorymocks "github.com/verasthiago/verancial/shared/repository/mocks"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestContext builds a gin.Context wired to an httptest.ResponseRecorder
// with the given JSON body already set on the request.
func newTestContext(t *testing.T, method, path string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	var reqBody []byte
	var err error
	switch b := body.(type) {
	case nil:
		reqBody = nil
	case []byte:
		reqBody = b
	case string:
		reqBody = []byte(b)
	default:
		reqBody, err = json.Marshal(b)
		require.NoError(t, err)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req

	return ctx, w
}

func hashedUser(t *testing.T, id, email, password string, isVerified, isAdmin bool) *models.User {
	t.Helper()
	user := &models.User{
		ID:         id,
		Email:      email,
		Name:       "Test User",
		IsVerified: isVerified,
		IsAdmin:    isAdmin,
	}
	require.NoError(t, user.HashPassword(password))
	return user
}

func newSharedFlags(jwtKey, jwtKeyEmail string) *sharedflags.SharedFlags {
	return &sharedflags.SharedFlags{
		JwtKey:      jwtKey,
		JwtKeyEmail: jwtKeyEmail,
	}
}

func TestLoginUserHandler_Handler(t *testing.T) {
	t.Run("valid credentials returns a JWT", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repositorymocks.NewMockRepository(ctrl)
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		user := hashedUser(t, "user-1", "jane@example.com", "correct-password", true, false)
		sharedFlags := newSharedFlags("jwt-secret", "jwt-email-secret")

		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockBuilder.EXPECT().GetSharedFlags().Return(sharedFlags).AnyTimes()
		mockRepo.EXPECT().GetUserByEmail("jane@example.com").Return(user, nil)

		handler := new(LoginUserHandler).InitFromBuilder(mockBuilder)

		ctx, w := newTestContext(t, http.MethodPost, "/login/v0/user/signin", map[string]string{
			"email":    "jane@example.com",
			"password": "correct-password",
		})

		err := handler.Handler(ctx)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "success", resp["status"])
		token, ok := resp["token"].(string)
		require.True(t, ok)
		assert.NotEmpty(t, token)
	})

	t.Run("user not found returns the repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repositorymocks.NewMockRepository(ctrl)
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		repoErr := errRecordNotFound()

		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockRepo.EXPECT().GetUserByEmail("missing@example.com").Return(nil, repoErr)

		handler := new(LoginUserHandler).InitFromBuilder(mockBuilder)

		ctx, _ := newTestContext(t, http.MethodPost, "/login/v0/user/signin", map[string]string{
			"email":    "missing@example.com",
			"password": "whatever",
		})

		err := handler.Handler(ctx)

		require.Error(t, err)
		assert.Equal(t, repoErr, err)
	})

	t.Run("unverified account returns UNVERIFIED_ACCOUNT error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repositorymocks.NewMockRepository(ctrl)
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		user := hashedUser(t, "user-2", "unverified@example.com", "correct-password", false, false)

		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockRepo.EXPECT().GetUserByEmail("unverified@example.com").Return(user, nil)

		handler := new(LoginUserHandler).InitFromBuilder(mockBuilder)

		ctx, _ := newTestContext(t, http.MethodPost, "/login/v0/user/signin", map[string]string{
			"email":    "unverified@example.com",
			"password": "correct-password",
		})

		err := handler.Handler(ctx)

		require.Error(t, err)
		assertGenericError(t, err, http.StatusUnauthorized, "UNVERIFIED_ACCOUNT", "Unverified account")
	})

	t.Run("wrong password returns INVALID_PASSWORD error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repositorymocks.NewMockRepository(ctrl)
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		user := hashedUser(t, "user-3", "jane@example.com", "correct-password", true, false)

		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockRepo.EXPECT().GetUserByEmail("jane@example.com").Return(user, nil)

		handler := new(LoginUserHandler).InitFromBuilder(mockBuilder)

		ctx, _ := newTestContext(t, http.MethodPost, "/login/v0/user/signin", map[string]string{
			"email":    "jane@example.com",
			"password": "wrong-password",
		})

		err := handler.Handler(ctx)

		require.Error(t, err)
		assertGenericError(t, err, http.StatusUnauthorized, "INVALID_PASSWORD", "Invalid password")
	})

	t.Run("request validation failure returns a validation error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		// GetRepository/GetUserByEmail must never be called because validation fails first.

		handler := new(LoginUserHandler).InitFromBuilder(mockBuilder)

		ctx, _ := newTestContext(t, http.MethodPost, "/login/v0/user/signin", map[string]string{
			"email":    "not-an-email",
			"password": "",
		})

		err := handler.Handler(ctx)

		require.Error(t, err)
		assertGenericError(t, err, http.StatusBadRequest, "INVALID_INPUT", "")
	})

	t.Run("malformed JSON body returns a bind error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		handler := new(LoginUserHandler).InitFromBuilder(mockBuilder)

		ctx, _ := newTestContext(t, http.MethodPost, "/login/v0/user/signin", "{not-json")

		err := handler.Handler(ctx)

		require.Error(t, err)
	})

	t.Run("CheckPassword non-mismatch bcrypt error propagates unchanged", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repositorymocks.NewMockRepository(ctrl)
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		// A user with an empty/invalid stored hash causes bcrypt to return
		// "hashedSecret too short to be a bcrypted password", which is NOT
		// bcrypt.ErrMismatchedHashAndPassword, so it should propagate as-is.
		user := &models.User{
			ID:         "user-4",
			Email:      "broken@example.com",
			Password:   "",
			IsVerified: true,
		}

		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockRepo.EXPECT().GetUserByEmail("broken@example.com").Return(user, nil)

		handler := new(LoginUserHandler).InitFromBuilder(mockBuilder)

		ctx, _ := newTestContext(t, http.MethodPost, "/login/v0/user/signin", map[string]string{
			"email":    "broken@example.com",
			"password": "whatever",
		})

		err := handler.Handler(ctx)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "hashedSecret too short")
	})
}

func TestLoginUserHandler_InitFromBuilder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilder := buildermocks.NewMockBuilder(ctrl)
	handler := new(LoginUserHandler).InitFromBuilder(mockBuilder)

	require.NotNil(t, handler)
	assert.Equal(t, builder.Builder(mockBuilder), handler.Builder)
}

// helper for tests that don't need a real logger but want to make sure
// GetLog wiring doesn't panic if exercised.
func discardLogger() *zap.Logger {
	return zap.NewNop()
}
