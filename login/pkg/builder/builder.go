package builder

import (
	shared "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/shared/repository"
	"go.uber.org/zap"
)

//go:generate mockgen -source=builder.go -destination=mocks/mock_builder.go -package=mocks

type Builder interface {
	GetRepository() repository.Repository
	GetFlags() *Flags
	GetLog() *zap.Logger
	GetSharedFlags() *shared.SharedFlags
	InitBuilder(loginEnvConfigFile, sharedEnvConfigFile *shared.EnvFileConfig) Builder
}
