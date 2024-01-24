package pkg

import (
	"github.com/gin-gonic/gin"
	"github.com/verasthiago/verancial/api/pkg/builder"
	"github.com/verasthiago/verancial/api/pkg/handlers"
	"github.com/verasthiago/verancial/api/pkg/middlewares"
	"github.com/verasthiago/verancial/shared/errors"
)

type Server struct {
	builder.Builder

	ReportProcessor handlers.ReportProcessorAPI
	AppIntegration  handlers.AppIntegrationAPI

	AuthAPI middlewares.AuthUserAPI
}

func (s *Server) InitFromBuilder(builder builder.Builder) *Server {
	s.Builder = builder
	s.ReportProcessor = new(handlers.ReportProcessorHandler).InitFromBuilder(builder)
	s.AppIntegration = new(handlers.AppIntegrationHandler).InitFromBuilder(builder)

	s.AuthAPI = new(middlewares.AuthUserHandler).InitFromFlags(builder.GetFlags(), builder.GetSharedFlags())
	return s
}

func (s *Server) Run() error {

	app := gin.Default()
	api := app.Group("/api")
	{
		apiV0 := api.Group("/v0").Use(s.AuthAPI.Handler())
		{
			apiV0.POST("report-processor", errors.ErrorRoute(s.ReportProcessor.Handler))
			apiV0.POST("app-integration", errors.ErrorRoute(s.AppIntegration.Handler))
		}
	}
	return app.Run(":" + s.GetFlags().Port)
}
