package pkg

import (
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/verasthiago/verancial/app-integration-worker/pkg/builder"
	"github.com/verasthiago/verancial/app-integration-worker/pkg/handlers"
	"github.com/verasthiago/verancial/shared/types"
)

type Server struct {
	builder.Builder

	AppIntegrationAPI handlers.AppIntegrationAPI
}

func (s *Server) InitFromBuilder(builder builder.Builder) *Server {
	s.Builder = builder

	s.AppIntegrationAPI = new(handlers.AppIntegrationHandler).InitFromBuilder(builder)
	return s
}

func (s *Server) Run() error {
	dsn := fmt.Sprintf("%+v:%+v", s.GetSharedFlags().QueueHost, s.GetSharedFlags().QueuePort)
	redisConnection := asynq.RedisClientOpt{
		Addr: dsn,
	}

	worker := asynq.NewServer(redisConnection, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
	})

	mux := asynq.NewServeMux()

	mux.HandleFunc(
		types.PatternAppIntegration,
		s.AppIntegrationAPI.Handler,
	)

	return worker.Run(mux)
}
