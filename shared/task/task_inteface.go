package task

import "github.com/verasthiago/verancial/shared/types"

type Task interface {
	CreateReportAsync(payload types.ReportProcessQueuePayload) error
	UpdateAppAsync(payload types.AppIntegrationQueuePayload) error
}
