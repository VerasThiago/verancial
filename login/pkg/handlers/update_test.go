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
	sharederrors "github.com/verasthiago/verancial/shared/errors"
	"github.com/verasthiago/verancial/shared/models"
	repositorymocks "github.com/verasthiago/verancial/shared/repository/mocks"
)

func TestUpdateUserHandler_Handler(t *testing.T) {
	t.Run("success updates the user identified by the body id and returns success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repositorymocks.NewMockRepository(ctrl)
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		targetID := uuid.New().String()

		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockRepo.EXPECT().UpdateUser(gomock.Any()).DoAndReturn(func(user *models.User) error {
			assert.Equal(t, targetID, user.ID)
			assert.Equal(t, "Janet", user.Name)
			return nil
		})

		handler := new(UpdateUserHandler).InitFromBuilder(mockBuilder)

		ctx, w := newTestContext(t, http.MethodPut, "/login/v0/admin/update", map[string]string{
			"id":   targetID,
			"name": "Janet",
		})

		err := handler.Handler(ctx)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "success", resp["status"])
	})

	// The handler itself performs no authorization/ownership check: it trusts
	// whatever "id" is present in the JSON body and updates that user. The
	// only gate is the AuthUserHandler admin middleware applied at the router
	// level (see pkg/server.go), which merely requires the caller's JWT to
	// have IsAdmin == true - it does NOT compare the JWT's user id against
	// the body's "id". This test documents that an admin can update an
	// arbitrary user id supplied in the body, since the handler has no
	// additional per-resource ownership check of its own.
	t.Run("handler applies no per-resource ownership check - any target id in the body is updated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repositorymocks.NewMockRepository(ctrl)
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		arbitraryOtherUserID := uuid.New().String()

		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockRepo.EXPECT().UpdateUser(gomock.Any()).DoAndReturn(func(user *models.User) error {
			assert.Equal(t, arbitraryOtherUserID, user.ID)
			return nil
		})

		handler := new(UpdateUserHandler).InitFromBuilder(mockBuilder)

		// No Authorization header / JWT claim is consulted anywhere in this
		// handler - the request body alone determines which user gets updated.
		ctx, w := newTestContext(t, http.MethodPut, "/login/v0/admin/update", map[string]string{
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

		handler := new(UpdateUserHandler).InitFromBuilder(mockBuilder)

		ctx, _ := newTestContext(t, http.MethodPut, "/login/v0/admin/update", map[string]string{
			"name": "Janet",
		})

		err := handler.Handler(ctx)

		require.Error(t, err)
		assertGenericError(t, err, http.StatusBadRequest, "INVALID_INPUT", "")
	})

	t.Run("validation failure (invalid uuid, non-alpha name, bad email)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		handler := new(UpdateUserHandler).InitFromBuilder(mockBuilder)

		ctx, _ := newTestContext(t, http.MethodPut, "/login/v0/admin/update", map[string]string{
			"id":    "not-a-uuid",
			"name":  "Janet123",
			"email": "not-an-email",
		})

		err := handler.Handler(ctx)

		require.Error(t, err)
		genericErr, ok := err.(sharederrors.GenericError)
		require.True(t, ok)
		assert.Contains(t, genericErr.Message, "valid UUID")
		assert.Contains(t, genericErr.Message, "may only contain letters")
		assert.Contains(t, genericErr.Message, "valid email address")
	})

	t.Run("repository error is propagated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repositorymocks.NewMockRepository(ctrl)
		mockBuilder := buildermocks.NewMockBuilder(ctrl)

		repoErr := errRecordNotFound()

		mockBuilder.EXPECT().GetRepository().Return(mockRepo).AnyTimes()
		mockRepo.EXPECT().UpdateUser(gomock.Any()).Return(repoErr)

		handler := new(UpdateUserHandler).InitFromBuilder(mockBuilder)

		ctx, _ := newTestContext(t, http.MethodPut, "/login/v0/admin/update", map[string]string{
			"id": uuid.New().String(),
		})

		err := handler.Handler(ctx)

		require.Error(t, err)
		assert.Equal(t, repoErr, err)
	})

	t.Run("malformed JSON body returns a bind error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		handler := new(UpdateUserHandler).InitFromBuilder(mockBuilder)

		ctx, _ := newTestContext(t, http.MethodPut, "/login/v0/admin/update", "{not-json")

		err := handler.Handler(ctx)

		require.Error(t, err)
	})
}

func TestUpdateUserHandler_InitFromBuilder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBuilder := buildermocks.NewMockBuilder(ctrl)
	handler := new(UpdateUserHandler).InitFromBuilder(mockBuilder)

	require.NotNil(t, handler)
}
