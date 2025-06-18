package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"encoding/base64"

	"github.com/gin-gonic/gin"
	"github.com/verasthiago/verancial/api/pkg/builder"
	"github.com/verasthiago/verancial/shared/constants"
	"github.com/verasthiago/verancial/shared/types"
)

type Request struct {
	UserId          string `json:"userid"`
	FilePath        string `json:"filepath,omitempty"` // Used internally for DPW service
	FileData        string `json:"filedata,omitempty"` // Base64 encoded CSV content from web uploads
	FileName        string `json:"filename,omitempty"` // Original filename from web uploads
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

	// Process the uploaded file
	tempFilePath, err := r.handleFileUpload(request)
	if err != nil {
		return err
	}

	if request.AsyncProcessing {
		return r.processAsync(request, tempFilePath, context)
	}

	return r.processSync(request, tempFilePath, context)
}

func (r *ReportProcessorHandler) handleFileUpload(request Request) (string, error) {
	if request.FileData == "" {
		return "", fmt.Errorf("file data is required for web uploads")
	}

	// Decode base64 file data
	fileData, err := base64.StdEncoding.DecodeString(request.FileData)
	if err != nil {
		return "", fmt.Errorf("failed to decode file data: %v", err)
	}

	// Create temporary file
	tempDir := os.TempDir()
	tempFilePath := filepath.Join(tempDir, fmt.Sprintf("upload_%s_%s", request.UserId, request.FileName))

	// Write decoded data to temporary file
	if err := ioutil.WriteFile(tempFilePath, fileData, 0644); err != nil {
		return "", fmt.Errorf("failed to write temporary file: %v", err)
	}

	return tempFilePath, nil
}

func (r *ReportProcessorHandler) processAsync(request Request, tempFilePath string, context *gin.Context) error {
	err := r.GetTask().CreateReportAsync(types.ReportProcessQueuePayload{
		UserId:   request.UserId,
		BankId:   constants.BankId(request.BankId),
		FilePath: tempFilePath,
	})
	if err != nil {
		return err
	}

	context.JSON(http.StatusOK, gin.H{"status": "queued"})
	return nil
}

func (r *ReportProcessorHandler) processSync(request Request, tempFilePath string, context *gin.Context) error {
	// Clean up temporary file after processing
	defer os.Remove(tempFilePath)

	// Prepare request for DPW service
	dpwRequest := Request{
		UserId:          request.UserId,
		FilePath:        tempFilePath,
		BankId:          request.BankId,
		AsyncProcessing: false,
	}

	jsonData, err := json.Marshal(dpwRequest)
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
