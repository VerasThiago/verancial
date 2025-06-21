package pkg

import (
	"github.com/gin-gonic/gin"
	"github.com/verasthiago/verancial/api/pkg/builder"
	"github.com/verasthiago/verancial/api/pkg/handlers"
	"github.com/verasthiago/verancial/api/pkg/handlers/transaction"
	"github.com/verasthiago/verancial/api/pkg/middlewares"
	"github.com/verasthiago/verancial/shared/errors"
)

type Server struct {
	builder.Builder

	ReportProcessorAPI handlers.ReportProcessorAPI
	AppIntegrationAPI  handlers.AppIntegrationAPI
	DashboardAPI       handlers.DashboardAPI
	BankAPI            handlers.BankStatsAPI

	ListTransactionsAPI transaction.ListTransactionsAPI

	AuthAPI middlewares.AuthUserAPI
}

func (s *Server) InitFromBuilder(builder builder.Builder) *Server {
	s.Builder = builder
	s.ReportProcessorAPI = new(handlers.ReportProcessorHandler).InitFromBuilder(builder)
	s.AppIntegrationAPI = new(handlers.AppIntegrationHandler).InitFromBuilder(builder)
	s.DashboardAPI = new(handlers.DashboardHandler).InitFromBuilder(builder)
	s.BankAPI = new(handlers.BankStatsHandler).InitFromBuilder(builder)
	s.ListTransactionsAPI = new(transaction.ListTransactionsHandler).InitFromBuilder(builder)

	s.AuthAPI = new(middlewares.AuthUserHandler).InitFromFlags(builder.GetFlags(), builder.GetSharedFlags())
	return s
}

func (s *Server) Run() error {

	app := gin.Default()
	api := app.Group("/api")
	{
		apiV0 := api.Group("/v0")
		{
			apiV0ReportProcessor := apiV0.Group("/report").Use(s.AuthAPI.Handler())
			{
				apiV0ReportProcessor.POST("/process", errors.ErrorRoute(s.ReportProcessorAPI.Handler))
			}

			apiV0AppIntegration := apiV0.Group("/app-integration").Use(s.AuthAPI.Handler())
			{
				apiV0AppIntegration.POST("generate", errors.ErrorRoute(s.AppIntegrationAPI.Handler))
			}

			apiV0Dashboard := apiV0.Group("/dashboard").Use(s.AuthAPI.Handler())
			{
				apiV0Dashboard.GET("/user", errors.ErrorRoute(s.DashboardAPI.GetUserDashboard))
			}

			apiV0Transaction := apiV0.Group("/transaction").Use(s.AuthAPI.Handler())
			{
				apiV0Transaction.GET("/list/:bankId", errors.ErrorRoute(s.ListTransactionsAPI.Handler))
			}

			apiV0Bank := apiV0.Group("/bank").Use(s.AuthAPI.Handler())
			{
				apiV0Bank.GET("/:bankId", errors.ErrorRoute(s.BankAPI.Handler))
			}
		}
	}
	return app.Run(":" + s.GetFlags().Port)
}
