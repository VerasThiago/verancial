package builder

import (
	"github.com/verasthiago/verancial/app-integration-worker/pkg/generators"
	shared "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/shared/repository"
	postgresrepository "github.com/verasthiago/verancial/shared/repository/postgresRepository"
)

type ServerBuilder struct {
	*shared.SharedFlags
	*Flags
	Repository                repository.Repository
	AppReportGeneratorFactory generators.AppReportGeneratorFactory
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

func (s *ServerBuilder) GetAppReportGeneratorFactory() generators.AppReportGeneratorFactory {
	return s.AppReportGeneratorFactory
}

func (s *ServerBuilder) InitBuilder(appIntegrationEnvConfigFile, sharedEnvFileConfig *shared.EnvFileConfig) Builder {
	flags, err := new(Flags).InitFromViper(appIntegrationEnvConfigFile)
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
	s.AppReportGeneratorFactory = generators.DefaultAppReportGeneratorFactory{}

	return s
}
