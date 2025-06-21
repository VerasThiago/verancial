package transaction

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/verasthiago/verancial/api/pkg/builder"
	"github.com/verasthiago/verancial/shared/models"
)

type ListTransactionsRequest struct {
	// Path parameter - must be a valid UUID
	BankId string `uri:"bankId" binding:"required,uuid"`

	// Query parameters with defaults
	Page     *int `form:"page" binding:"omitempty,min=1"`
	PageSize *int `form:"pageSize" binding:"omitempty,min=1,max=100"`

	// Optional filters
	Uncategorized *bool `form:"uncategorized"`
}

type ListTransactionsAPI interface {
	Handler(c *gin.Context) error
}

type ListTransactionsHandler struct {
	builder.Builder
}

func (h *ListTransactionsHandler) InitFromBuilder(builder builder.Builder) *ListTransactionsHandler {
	h.Builder = builder
	return h
}

func (h *ListTransactionsHandler) Handler(c *gin.Context) error {
	userObj, _ := c.Get("user")
	user, _ := userObj.(*models.User)

	var req ListTransactionsRequest
	if err := c.ShouldBindUri(&req); err != nil {
		return fmt.Errorf("invalid bank ID (must be a valid UUID): %v", err)
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		return fmt.Errorf("invalid query parameters: %v", err)
	}

	// Set defaults if not provided
	defaultPage := 1
	if req.Page == nil {
		req.Page = &defaultPage
	}

	defaultPageSize := 50
	if req.PageSize == nil {
		req.PageSize = &defaultPageSize
	}

	defaultUncategorized := false
	if req.Uncategorized == nil {
		req.Uncategorized = &defaultUncategorized
	}

	// Calculate offset from page and pageSize
	offset := (*req.Page - 1) * *req.PageSize

	transactions, err := h.GetRepository().GetTransactions(
		user.ID,
		string(req.BankId),
		*req.PageSize,
		offset,
		&models.TransactionFilter{
			Uncategorized: *req.Uncategorized,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to get transactions: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   transactions,
	})
	return nil
}
