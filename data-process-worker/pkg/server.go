package pkg

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/verasthiago/verancial/data-process-worker/pkg/builder"
	"github.com/verasthiago/verancial/data-process-worker/pkg/handlers"
	"github.com/verasthiago/verancial/shared/errors"
	"github.com/verasthiago/verancial/shared/types"
)

type Server struct {
	builder.Builder

	ReportCreateAPI handlers.CreateReportAPI
}

func (s *Server) InitFromBuilder(builder builder.Builder) *Server {
	s.Builder = builder

	s.ReportCreateAPI = new(handlers.CreateReportHandler).InitFromBuilder(builder)
	return s
}

func (s *Server) RunAsync() error {
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
		types.PatternReportProcess,
		s.ReportCreateAPI.HandlerAsync,
	)

	return worker.Run(mux)
}

func (s *Server) RunSync() error {
	app := gin.Default()
	api := app.Group("/dpw")
	{
		apiV0 := api.Group("/v0")
		{
			apiV0.POST("process_report", errors.ErrorRoute(s.ReportCreateAPI.HandlerSync))
		}
	}

	return app.Run(":" + s.GetFlags().Port)
}

func (s *Server) Run() error {
	if s.GetFlags().AsyncProcessing {
		return s.RunAsync()
	}

	return s.RunSync()
}
