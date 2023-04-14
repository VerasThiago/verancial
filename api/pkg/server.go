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

	Hello handlers.HelloAPI

	AuthAPI middlewares.AuthUserAPI
}

func (s *Server) InitFromBuilder(builder builder.Builder) *Server {
	s.Builder = builder
	s.Hello = new(handlers.HelloHandler).InitFromBuilder(builder)

	s.AuthAPI = new(middlewares.AuthUserHandler).InitFromFlags(builder.GetFlags(), builder.GetSharedFlags())
	return s
}

func (s *Server) Run() error {

	app := gin.Default()
	api := app.Group("/api")
	{
		apiV0 := api.Group("/v0").Use(s.AuthAPI.Handler())
		{
			apiV0.GET("hello", errors.ErrorRoute(s.Hello.Handler))
		}
	}
	return app.Run(":" + s.GetFlags().Port)
}
