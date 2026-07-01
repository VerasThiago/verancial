package builder

import (
	"github.com/verasthiago/verancial/app-integration-worker/pkg/generators"
	shared "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/shared/repository"
)

//go:generate mockgen -source=builder.go -destination=mocks/mock_builder.go -package=mocks

type Builder interface {
	GetRepository() repository.Repository
	GetSharedFlags() *shared.SharedFlags
	GetFlags() *Flags
	GetAppReportGeneratorFactory() generators.AppReportGeneratorFactory
	InitBuilder(appIntegrationEnvConfigFile, sharedEnvFileConfig *shared.EnvFileConfig) Builder
}
