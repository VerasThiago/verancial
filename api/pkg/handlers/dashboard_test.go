package handlers

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	buildermocks "github.com/verasthiago/verancial/api/pkg/builder/mocks"
	"github.com/verasthiago/verancial/api/pkg/testutil"
	sharedflags "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/shared/models"
	repomocks "github.com/verasthiago/verancial/shared/repository/mocks"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardHandler_GetUserDashboard(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid JWT returns dashboard stats as JSON", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockRepo := repomocks.NewMockRepository(ctrl)

		userID := "user-123"
		token := testutil.GenerateToken(t, testutil.TestJwtKey, userID, false, time.Now().Add(time.Hour))

		stats := &models.UserDashboardStats{
			TotalBankAccounts: 2,
			BankAccountStats:  []models.BankAccountStat{},
		}

		mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{JwtKey: testutil.TestJwtKey}).AnyTimes()
		mockBuilder.EXPECT().GetRepository().Return(mockRepo)
		mockRepo.EXPECT().GetUserDashboardStats(userID).Return(stats, nil)

		c, w := testutil.NewGinContext("GET", "/api/v0/dashboard/user", nil, token)

		handler := new(DashboardHandler).InitFromBuilder(mockBuilder)
		err := handler.GetUserDashboard(c)

		require.NoError(t, err)
		assert.Equal(t, 200, w.Code)

		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, "success", body["status"])
		data := body["data"].(map[string]interface{})
		assert.Equal(t, float64(2), data["total_bank_accounts"])
	})

	t.Run("missing JWT returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{JwtKey: testutil.TestJwtKey}).AnyTimes()

		c, _ := testutil.NewGinContext("GET", "/api/v0/dashboard/user", nil, "")

		handler := new(DashboardHandler).InitFromBuilder(mockBuilder)
		err := handler.GetUserDashboard(c)

		require.Error(t, err)
	})

	t.Run("invalid JWT returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{JwtKey: testutil.TestJwtKey}).AnyTimes()

		c, _ := testutil.NewGinContext("GET", "/api/v0/dashboard/user", nil, "not-a-real-token")

		handler := new(DashboardHandler).InitFromBuilder(mockBuilder)
		err := handler.GetUserDashboard(c)

		require.Error(t, err)
	})

	t.Run("expired JWT returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{JwtKey: testutil.TestJwtKey}).AnyTimes()

		userID := "user-123"
		token := testutil.GenerateExpiredToken(t, userID)

		c, _ := testutil.NewGinContext("GET", "/api/v0/dashboard/user", nil, token)

		handler := new(DashboardHandler).InitFromBuilder(mockBuilder)
		err := handler.GetUserDashboard(c)

		// Verified empirically: jwt.ParseWithClaims invokes Claims.Valid(),
		// and jwt-go's default StandardClaims.Valid() rejects an expired
		// exp, so GetJWTClaimFromToken itself errors out before the
		// repository would ever be reached.
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("repository error is propagated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockRepo := repomocks.NewMockRepository(ctrl)

		userID := "user-123"
		token := testutil.GenerateToken(t, testutil.TestJwtKey, userID, false, time.Now().Add(time.Hour))

		mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{JwtKey: testutil.TestJwtKey}).AnyTimes()
		mockBuilder.EXPECT().GetRepository().Return(mockRepo)
		mockRepo.EXPECT().GetUserDashboardStats(userID).Return(nil, fmt.Errorf("db exploded"))

		c, _ := testutil.NewGinContext("GET", "/api/v0/dashboard/user", nil, token)

		handler := new(DashboardHandler).InitFromBuilder(mockBuilder)
		err := handler.GetUserDashboard(c)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "db exploded")
	})

	t.Run("wrong signing key rejected", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{JwtKey: testutil.TestJwtKey}).AnyTimes()

		token := testutil.GenerateToken(t, "a-different-key", "user-123", false, time.Now().Add(time.Hour))
		c, _ := testutil.NewGinContext("GET", "/api/v0/dashboard/user", nil, token)

		handler := new(DashboardHandler).InitFromBuilder(mockBuilder)
		err := handler.GetUserDashboard(c)

		require.Error(t, err)
	})

	t.Run("none-alg token rejected", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBuilder := buildermocks.NewMockBuilder(ctrl)
		mockBuilder.EXPECT().GetSharedFlags().Return(&sharedflags.SharedFlags{JwtKey: testutil.TestJwtKey}).AnyTimes()

		token := testutil.GenerateTokenSignedWithNone(t, "user-123")
		c, _ := testutil.NewGinContext("GET", "/api/v0/dashboard/user", nil, token)

		handler := new(DashboardHandler).InitFromBuilder(mockBuilder)
		err := handler.GetUserDashboard(c)

		require.Error(t, err)
	})
}
