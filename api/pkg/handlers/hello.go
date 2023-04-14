package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/verasthiago/verancial/api/pkg/builder"
)

type HelloAPI interface {
	Handler(context *gin.Context) error
}

type HelloHandler struct {
	builder.Builder
}

func (l *HelloHandler) InitFromBuilder(builder builder.Builder) *HelloHandler {
	l.Builder = builder
	return l
}
func (l *HelloHandler) Handler(context *gin.Context) error {
	context.JSON(http.StatusOK, gin.H{"status": "success", "message": "hello!"})
	return nil
}
