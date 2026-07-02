package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	buildermocks "github.com/verasthiago/verancial/login/pkg/builder/mocks"
	repositorymocks "github.com/verasthiago/verancial/shared/repository/mocks"
)

func TestDeleteUserHandler_Handler(t *testing.T) {
	t.Run("success deletes the user identified by the body id and returns success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repositorymocks.NewMockRepository(ctrl)
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		targetID := uuid.New().String()

		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockRepo.EXPECT().DeleteUser(targetID).Return(nil)

		handler := new(DeleteUserHandler).InitFromBuilder(mockBuilder)

		ctx, w := newTestContext(t, http.MethodDelete, "/login/v0/admin/delete", map[string]string{
			"id": targetID,
		})

		err := handler.Handler(ctx)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "success", resp["status"])
	})

	// Like UpdateUserHandler, DeleteUserHandler performs no per-resource
	// ownership check: it deletes whatever "id" is present in the JSON body.
	// The only gate is the router-level admin middleware (must be an admin
	// JWT), not "must be deleting yourself". This documents that behavior.
	t.Run("handler applies no per-resource ownership check - any target id in the body is deleted", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repositorymocks.NewMockRepository(ctrl)
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		arbitraryOtherUserID := uuid.New().String()

		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockRepo.EXPECT().DeleteUser(arbitraryOtherUserID).Return(nil)

		handler := new(DeleteUserHandler).InitFromBuilder(mockBuilder)

		// No Authorization header / JWT claim is read anywhere in this handler.
		ctx, w := newTestContext(t, http.MethodDelete, "/login/v0/admin/delete", map[string]string{
			"id": arbitraryOtherUserID,
		})

		err := handler.Handler(ctx)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("validation failure (missing id) prevents repository call", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		// No repository expectations: any call fails the test via gomock.

		handler := new(DeleteUserHandler).InitFromBuilder(mockBuilder)

		ctx, _ := newTestContext(t, http.MethodDelete, "/login/v0/admin/delete", map[string]string{})

		err := handler.Handler(ctx)

		require.Error(t, err)
		assertGenericError(t, err, http.StatusBadRequest, "INVALID_INPUT", "The id field is required")
	})

	t.Run("validation failure (invalid uuid)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		handler := new(DeleteUserHandler).InitFromBuilder(mockBuilder)

		ctx, _ := newTestContext(t, http.MethodDelete, "/login/v0/admin/delete", map[string]string{
			"id": "not-a-uuid",
		})

		err := handler.Handler(ctx)

		require.Error(t, err)
		assertGenericError(t, err, http.StatusBadRequest, "INVALID_INPUT", "The id field must contain valid UUID")
	})

	t.Run("repository error is propagated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repositorymocks.NewMockRepository(ctrl)
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		targetID := uuid.New().String()
		repoErr := errRecordNotFound()

		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockRepo.EXPECT().DeleteUser(targetID).Return(repoErr)

		handler := new(DeleteUserHandler).InitFromBuilder(mockBuilder)

		ctx, _ := newTestContext(t, http.MethodDelete, "/login/v0/admin/delete", map[string]string{
			"id": targetID,
		})

		err := handler.Handler(ctx)

		require.Error(t, err)
		assert.Equal(t, repoErr, err)
	})

	t.Run("malformed JSON body returns a bind error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		handler := new(DeleteUserHandler).InitFromBuilder(mockBuilder)

		ctx, _ := newTestContext(t, http.MethodDelete, "/login/v0/admin/delete", "{not-json")

		err := handler.Handler(ctx)

		require.Error(t, err)
	})
}

func TestDeleteUserHandler_InitFromBuilder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilder := buildermocks.NewMockBuilder(ctrl)
	handler := new(DeleteUserHandler).InitFromBuilder(mockBuilder)

	require.NotNil(t, handler)
}
