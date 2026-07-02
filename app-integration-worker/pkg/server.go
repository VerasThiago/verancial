package pkg

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/verasthiago/verancial/app-integration-worker/pkg/builder"
	"github.com/verasthiago/verancial/app-integration-worker/pkg/handlers"
	"github.com/verasthiago/verancial/shared/errors"
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
	if s.GetFlags().AsyncProcessing {
		return s.RunAsync()
	}

	return s.RunSync()
}

// SetupRouter builds the gin engine with every route registered, without
// binding a listener. Split out from RunSync so route wiring can be
// exercised with httptest instead of a real network listener.
func (s *Server) SetupRouter() *gin.Engine {
	app := gin.Default()
	api := app.Group("/aiw")
	{
		apiV0 := api.Group("/v0")
		{
			apiV0.POST("process_app_report", errors.ErrorRoute(s.AppIntegrationAPI.HandlerSync))
		}
	}
	return app
}

func (s *Server) RunSync() error {
	return s.SetupRouter().Run(":" + s.GetFlags().Port)
}

// SetupAsyncWorker builds the asynq server and mux with every task handler
// registered, without starting the (blocking) processing loop. Split out
// from RunAsync so the wiring -- which queue config, which task pattern
// maps to which handler -- can be exercised without a real Redis and
// without blocking forever.
func (s *Server) SetupAsyncWorker() (*asynq.Server, *asynq.ServeMux) {
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
		s.AppIntegrationAPI.HandlerAsync,
	)

	return worker, mux
}

func (s *Server) RunAsync() error {
	worker, mux := s.SetupAsyncWorker()
	return worker.Run(mux)
}
