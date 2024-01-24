package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/verasthiago/verancial/api/pkg/builder"
	"github.com/verasthiago/verancial/shared/constants"
	"github.com/verasthiago/verancial/shared/types"
)

type ReportProcessorAPI interface {
	Handler(context *gin.Context) error
}

type ReportProcessorHandler struct {
	builder.Builder
}

func (l *ReportProcessorHandler) InitFromBuilder(builder builder.Builder) *ReportProcessorHandler {
	l.Builder = builder
	return l
}
func (l *ReportProcessorHandler) Handler(context *gin.Context) error {
	var err error
	var request struct {
		UserId   string `json:"userid"`
		FilePath string `json:"filepath"`
		BankId   string `json:"bankid"`
	}

	if err := context.ShouldBindJSON(&request); err != nil {
		return err
	}

	fmt.Printf("\n ---- request %+v ---- \n", request)
	if err = l.GetTask().CreateReportAsync(types.ReportProcessQueuePayload{
		UserId:   request.UserId,
		BankId:   constants.BankId(request.BankId),
		FilePath: request.FilePath,
	}); err != nil {
		return err
	}

	context.JSON(http.StatusOK, gin.H{"status": "queued"})
	return nil
}
