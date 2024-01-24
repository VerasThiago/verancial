package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/verasthiago/verancial/api/pkg/builder"
	"github.com/verasthiago/verancial/shared/constants"
	"github.com/verasthiago/verancial/shared/types"
)

type AppIntegrationAPI interface {
	Handler(context *gin.Context) error
}

type AppIntegrationHandler struct {
	builder.Builder
}

func (l *AppIntegrationHandler) InitFromBuilder(builder builder.Builder) *AppIntegrationHandler {
	l.Builder = builder
	return l
}
func (l *AppIntegrationHandler) Handler(context *gin.Context) error {
	var err error
	var request struct {
		UserId              string `json:"userid"`
		AppID               string `json:"appid"`
		BankId              string `json:"bankid"`
		LastTransactionData string `json:"lasttransactiondate"`
	}

	if err := context.ShouldBindJSON(&request); err != nil {
		return err
	}
	fmt.Printf("\nrequest %+v\n", request)

	if err = l.GetTask().UpdateAppAsync(types.AppIntegrationQueuePayload{
		UserId:              request.UserId,
		AppID:               constants.AppID(request.AppID),
		BankId:              constants.BankId(request.BankId),
		LastTransactionDate: request.LastTransactionData,
	}); err != nil {
		return err
	}

	context.JSON(http.StatusOK, gin.H{"status": "queued"})
	return nil
}
