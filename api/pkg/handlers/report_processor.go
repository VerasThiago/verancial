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

type Request struct {
	UserId          string `json:"userid"`
	FilePath        string `json:"filepath"`
	BankId          string `json:"bankid"`
	AsyncProcessing bool   `json:"asyncprocessing"`
}

type ReportProcessorAPI interface {
	Handler(context *gin.Context) error
}

type ReportProcessorHandler struct {
	builder.Builder
}

func (r *ReportProcessorHandler) InitFromBuilder(builder builder.Builder) *ReportProcessorHandler {
	r.Builder = builder
	return r
}

func (r *ReportProcessorHandler) Handler(context *gin.Context) error {
	var request Request

	if err := context.ShouldBindJSON(&request); err != nil {
		return err
	}

	fmt.Printf("\n[API] Request: %+v\n\n", request)

	if request.AsyncProcessing {
		return r.SendAsyncRequest(request, context)
	}

	return r.SendSyncRequest(request, context)
}

func (r *ReportProcessorHandler) SendAsyncRequest(request Request, context *gin.Context) error {
	var err error

	if err = r.GetTask().CreateReportAsync(types.ReportProcessQueuePayload{
		UserId:   request.UserId,
		BankId:   constants.BankId(request.BankId),
		FilePath: request.FilePath,
	}); err != nil {
		return err
	}

	context.JSON(http.StatusOK, gin.H{"status": "queued"})
	return nil
}

func (r *ReportProcessorHandler) SendSyncRequest(request Request, context *gin.Context) error {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%v:%v/dpw/v0/process_report", r.GetSharedFlags().DPWHost, r.GetSharedFlags().DPWPort)
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
		return fmt.Errorf("received status code %d from [DPW] service", resp.StatusCode)
	}

	context.JSON(resp.StatusCode, gin.H{
		"status": "Report processed successfully",
	})

	return nil
}
