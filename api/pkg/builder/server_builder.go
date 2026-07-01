package builder

import (
	"net/http"
	"time"

	shared "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/shared/httpclient"
	"github.com/verasthiago/verancial/shared/repository"
	postgresrepository "github.com/verasthiago/verancial/shared/repository/postgresRepository"
	"github.com/verasthiago/verancial/shared/task"
	"go.uber.org/zap"
)

type ServerBuilder struct {
	*Flags
	*shared.SharedFlags
	Repository repository.Repository
	Task       task.Task
	HTTPClient httpclient.HTTPClient
	Log        *zap.Logger
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

func (s *ServerBuilder) GetLog() *zap.Logger {
	return s.Log
}

func (s *ServerBuilder) GetTask() task.Task {
	return s.Task
}

func (s *ServerBuilder) GetHTTPClient() httpclient.HTTPClient {
	return s.HTTPClient
}

func (s *ServerBuilder) InitBuilder(apiEnvFileConfig, sharedEnvFileConfig *shared.EnvFileConfig) Builder {
	flags, err := new(Flags).InitFromViper(apiEnvFileConfig)
	if err != nil {
		panic(err)
	}
	s.Flags = flags

	sharedFlags, err := new(shared.SharedFlags).InitFromViper(sharedEnvFileConfig)
	if err != nil {
		panic(err)
	}

	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	s.SharedFlags = sharedFlags
	s.Log = log
	s.Task = new(task.AsyncQueue).InitFromFlags(s.SharedFlags)
	s.Repository = new(postgresrepository.PostgresRepository).InitFromFlags(s.SharedFlags)
	s.HTTPClient = httpclient.New(&http.Client{Timeout: 5 * time.Minute})

	return s
}
