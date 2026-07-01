package task

import "github.com/verasthiago/verancial/shared/types"

//go:generate mockgen -source=task_inteface.go -destination=mocks/mock_task.go -package=mocks

type Task interface {
	CreateReportAsync(payload types.ReportProcessQueuePayload) error
	UpdateAppAsync(payload types.AppIntegrationQueuePayload) error
}
