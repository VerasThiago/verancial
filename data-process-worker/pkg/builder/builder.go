package builder

import (
	"github.com/verasthiago/verancial/data-process-worker/pkg/report"
	shared "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/shared/repository"
)

//go:generate mockgen -source=builder.go -destination=mocks/mock_builder.go -package=mocks

type Builder interface {
	GetRepository() repository.Repository
	GetSharedFlags() *shared.SharedFlags
	GetFlags() *Flags
	GetReportProcessorFactory() report.ReportProcessorFactory
	InitBuilder(dataProcessEnvConfigFile, sharedEnvFileConfig *shared.EnvFileConfig) Builder
}
