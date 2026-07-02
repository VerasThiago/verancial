package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	buildermocks "github.com/verasthiago/verancial/login/pkg/builder/mocks"
	sharederrors "github.com/verasthiago/verancial/shared/errors"
	"github.com/verasthiago/verancial/shared/models"
	repositorymocks "github.com/verasthiago/verancial/shared/repository/mocks"
)

func TestCreateUserHandler_Handler(t *testing.T) {
	t.Run("successful signup hashes the password before calling CreateUser and returns a token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repositorymocks.NewMockRepository(ctrl)
		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		sharedFlags := newSharedFlags("jwt-secret", "jwt-email-secret")

		const plaintextPassword = "super-secret-password"

		var capturedUser *models.User
		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockBuilder.EXPECT().GetSharedFlags().Return(sharedFlags).AnyTimes()
		mockRepo.EXPECT().CreateUser(gomock.Any()).DoAndReturn(func(user *models.User) error {
			capturedUser = user
			return nil
		})

		handler := new(CreateUserHandler).InitFromBuilder(mockBuilder)

		ctx, w := newTestContext(t, http.MethodPost, "/login/v0/user/signup", map[string]string{
			"name":     "Jane Doe",
			"email":    "jane@example.com",
			"password": plaintextPassword,
		})

		err := handler.Handler(ctx)

		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)

		require.NotNil(t, capturedUser)
		// The password stored/sent to the repository must NOT be the plaintext password...
		assert.NotEqual(t, plaintextPassword, capturedUser.Password)
		// ...and must be a valid bcrypt hash of the plaintext password.
		assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(capturedUser.Password), []byte(plaintextPassword)))

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Contains(t, resp, "id")
		token, ok := resp["token"].(string)
		require.True(t, ok)
		assert.NotEmpty(t, token)
	})

	t.Run("duplicate email repository error is propagated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repositorymocks.NewMockRepository(ctrl)
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		duplicateErr := sharederrors.GenericError{
			Code:    sharederrors.STATUS_BAD_REQUEST,
			Type:    sharederrors.DATA_ALREADY_BEGIN_USED.Type,
			Message: sharederrors.DATA_ALREADY_BEGIN_USED.Message,
		}

		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockRepo.EXPECT().CreateUser(gomock.Any()).Return(duplicateErr)

		handler := new(CreateUserHandler).InitFromBuilder(mockBuilder)

		ctx, _ := newTestContext(t, http.MethodPost, "/login/v0/user/signup", map[string]string{
			"name":     "Jane Doe",
			"email":    "jane@example.com",
			"password": "super-secret-password",
		})

		err := handler.Handler(ctx)

		require.Error(t, err)
		assertGenericError(t, err, http.StatusBadRequest, "DATA_ALREADY_BEGIN_USED", "Data is already being used")
	})

	t.Run("validation failure prevents repository call", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		// No GetRepository/CreateUser expectations set: any call fails the test.

		handler := new(CreateUserHandler).InitFromBuilder(mockBuilder)

		ctx, _ := newTestContext(t, http.MethodPost, "/login/v0/user/signup", map[string]string{
			"name":     "",
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
		handler := new(CreateUserHandler).InitFromBuilder(mockBuilder)

		ctx, _ := newTestContext(t, http.MethodPost, "/login/v0/user/signup", "{not-json")

		err := handler.Handler(ctx)

		require.Error(t, err)
	})
}

func TestCreateUserHandler_InitFromBuilder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilder := buildermocks.NewMockBuilder(ctrl)
	handler := new(CreateUserHandler).InitFromBuilder(mockBuilder)

	require.NotNil(t, handler)
}
