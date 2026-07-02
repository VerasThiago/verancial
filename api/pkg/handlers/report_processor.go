package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"encoding/base64"

	"github.com/gin-gonic/gin"
	"github.com/verasthiago/verancial/api/pkg/builder"
	"github.com/verasthiago/verancial/shared/auth"
	"github.com/verasthiago/verancial/shared/types"
)

// unsafeFileNameChars matches anything that isn't alphanumeric, dot, dash or
// underscore. Used to strip path separators / traversal sequences out of
// client-supplied file names before they are used to build a filesystem path.
var unsafeFileNameChars = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

// sanitizeFileName strips directory components and any character that isn't
// safe for a flat file name, preventing path traversal (e.g. "../../etc/x")
// when building the temp file path.
func sanitizeFileName(name string) string {
	name = filepath.Base(name)
	name = unsafeFileNameChars.ReplaceAllString(name, "_")
	if name == "" || name == "." || name == ".." {
		name = "upload"
	}
	return name
}

type Request struct {
	// UserId is intentionally not read from client input. It is derived
	// from the authenticated JWT to prevent IDOR (a user uploading or
	// processing a report on another user's behalf).
	UserId          string `json:"-"`
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

	userObj, _ := context.Get("user")
	user, ok := userObj.(*auth.UserClaims)
	if !ok || user == nil || user.ID == "" {
		return fmt.Errorf("unauthorized")
	}
	request.UserId = user.ID

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

	// Create temporary file. Both the user ID (from the authenticated JWT)
	// and the file name (from client input) are sanitized to plain,
	// flat-path-safe tokens before being joined onto tempDir, preventing
	// path traversal via a crafted "filename" (e.g. "../../etc/passwd").
	tempDir := os.TempDir()
	safeUserId := sanitizeFileName(request.UserId)
	safeFileName := sanitizeFileName(request.FileName)
	tempFilePath := filepath.Join(tempDir, fmt.Sprintf("upload_%s_%s", safeUserId, safeFileName))

	// Defense in depth: ensure the resolved path is still inside tempDir.
	if rel, err := filepath.Rel(tempDir, tempFilePath); err != nil || rel == ".." || len(rel) >= 2 && rel[:2] == ".." {
		return "", fmt.Errorf("invalid file path")
	}

	// Write decoded data to temporary file
	if err := ioutil.WriteFile(tempFilePath, fileData, 0600); err != nil {
		return "", fmt.Errorf("failed to write temporary file: %v", err)
	}

	return tempFilePath, nil
}

func (r *ReportProcessorHandler) processAsync(request Request, tempFilePath string, context *gin.Context) error {
	err := r.GetTask().CreateReportAsync(types.ReportProcessQueuePayload{
		UserId:   request.UserId,
		BankId:   string(request.BankId),
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

	// Prepare request for DPW service. Built from types.ReportProcessQueuePayload
	// (not the local Request type) because Request.UserId is tagged `json:"-"`
	// -- correct for the inbound client request (prevents a spoofed UserId in
	// the JSON body), but it would silently drop UserId from this outbound
	// call to DPW if reused here.
	dpwRequest := types.ReportProcessQueuePayload{
		UserId:   request.UserId,
		FilePath: tempFilePath,
		BankId:   request.BankId,
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

	resp, err := r.GetHTTPClient().Do(req)
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
