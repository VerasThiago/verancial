package builder

import (
	shared "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/shared/repository"
)

type Builder interface {
	GetRepository() repository.Repository
	GetSharedFlags() *shared.SharedFlags
	InitBuilder(sharedEnvFileConfig *shared.EnvFileConfig) Builder
}
