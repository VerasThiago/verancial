package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/verasthiago/verancial/api/pkg/builder"
	"github.com/verasthiago/verancial/shared/auth"
)

type DashboardAPI interface {
	GetUserDashboard(context *gin.Context) error
}

type DashboardHandler struct {
	builder.Builder
}

func (d *DashboardHandler) InitFromBuilder(builder builder.Builder) *DashboardHandler {
	d.Builder = builder
	return d
}

func (d *DashboardHandler) GetUserDashboard(context *gin.Context) error {
	jwtClaim, err := auth.GetJWTClaimFromToken(context.GetHeader("Authorization"), d.GetSharedFlags().JwtKey)
	if err != nil {
		return err
	}

	stats, err := d.GetRepository().GetUserDashboardStats(jwtClaim.User.ID)
	if err != nil {
		return err
	}

	context.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
	return nil
}
