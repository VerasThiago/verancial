package builder

import (
	flags "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/shared/httpclient"
	"github.com/verasthiago/verancial/shared/repository"
	task "github.com/verasthiago/verancial/shared/task"
	"go.uber.org/zap"
)

//go:generate mockgen -source=builder.go -destination=mocks/mock_builder.go -package=mocks

type Builder interface {
	GetRepository() repository.Repository
	GetFlags() *Flags
	GetLog() *zap.Logger
	GetSharedFlags() *flags.SharedFlags
	GetTask() task.Task
	GetHTTPClient() httpclient.HTTPClient
	InitBuilder(apiEnvConfigFile, sharedEnvConfigFile *flags.EnvFileConfig) Builder
}
