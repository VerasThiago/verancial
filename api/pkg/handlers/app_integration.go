package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/verasthiago/verancial/api/pkg/builder"
	"github.com/verasthiago/verancial/shared/constants"
	"github.com/verasthiago/verancial/shared/types"
)

type AppIntegrationRequest struct {
	UserId              string `json:"userid"`
	AppID               string `json:"appid"`
	BankId              string `json:"bankid"`
	LastTransactionData string `json:"lasttransactiondate"`
	AsyncProcessing     bool   `json:"asyncprocessing"`
}

type AppIntegrationAPI interface {
	Handler(context *gin.Context) error
}

type AppIntegrationHandler struct {
	builder.Builder
}

func (a *AppIntegrationHandler) InitFromBuilder(builder builder.Builder) *AppIntegrationHandler {
	a.Builder = builder
	return a
}

func (a *AppIntegrationHandler) Handler(context *gin.Context) error {
	var request AppIntegrationRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		return err
	}
	fmt.Printf("\nrequest %+v\n", request)

	if request.AsyncProcessing {
		return a.HandlerAsync(request, context)
	}

	return a.HandlerSync(request, context)
}

func (a *AppIntegrationHandler) HandlerSync(request AppIntegrationRequest, context *gin.Context) error {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%v:%v/aiw/v0/process_app_report", a.GetSharedFlags().AIWHost, a.GetSharedFlags().AIWPort)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received status code %d from [AIW] service", resp.StatusCode)
	}

	context.JSON(resp.StatusCode, gin.H{
		"status": "App report generated successfully",
	})

	return nil
}

func (a *AppIntegrationHandler) HandlerAsync(request AppIntegrationRequest, context *gin.Context) error {
	if err := a.GetTask().UpdateAppAsync(types.AppIntegrationQueuePayload{
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
