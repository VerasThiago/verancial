package builder

import (
	shared "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/shared/repository"
	postgresrepository "github.com/verasthiago/verancial/shared/repository/postgresRepository"
)

type ServerBuilder struct {
	*Flags
	*shared.SharedFlags
	Repository repository.Repository
}

func (s *ServerBuilder) GetFlags() *Flags {
	return s.Flags
}

func (s *ServerBuilder) GetSharedFlags() *shared.SharedFlags {
	return s.SharedFlags
}

func (s *ServerBuilder) GetRepository() repository.Repository {
	return s.Repository
}

func (s *ServerBuilder) InitBuilder(dataProcessEnvConfigFile, sharedEnvFileConfig *shared.EnvFileConfig) Builder {
	flags, err := new(Flags).InitFromViper(dataProcessEnvConfigFile)
	if err != nil {
		panic(err)
	}
	s.Flags = flags

	sharedflags, err := new(shared.SharedFlags).InitFromViper(sharedEnvFileConfig)
	if err != nil {
		panic(err)
	}

	s.SharedFlags = sharedflags
	s.Repository = new(postgresrepository.PostgresRepository).InitFromFlags(s.SharedFlags)

	return s
}
