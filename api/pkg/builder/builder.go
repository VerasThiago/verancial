package builder

import (
	flags "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/shared/repository"
	task "github.com/verasthiago/verancial/shared/task"
	"go.uber.org/zap"
)

type Builder interface {
	GetRepository() repository.Repository
	GetFlags() *Flags
	GetLog() *zap.Logger
	GetSharedFlags() *flags.SharedFlags
	GetTask() task.Task
	InitBuilder(apiEnvConfigFile, sharedEnvConfigFile *flags.EnvFileConfig) Builder
}
