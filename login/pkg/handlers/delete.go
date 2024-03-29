package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/verasthiago/verancial/login/pkg/builder"
	"github.com/verasthiago/verancial/login/pkg/validator"
	"github.com/verasthiago/verancial/shared/errors"
)

type DeleteUserAPI interface {
	Handler(context *gin.Context) error
}

type DeleteUserHandler struct {
	builder.Builder
}

func (l *DeleteUserHandler) InitFromBuilder(builder builder.Builder) *DeleteUserHandler {
	l.Builder = builder
	return l
}

func (l *DeleteUserHandler) Handler(context *gin.Context) error {
	var request validator.DeleteRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		return err
	}

	if errList := request.Validate(); len(errList) > 0 {
		return errors.CreateGenericErrorFromValidateError(errList)
	}

	if err := l.GetRepository().DeleteUser(request.UserId); err != nil {
		return err
	}

	context.JSON(http.StatusOK, gin.H{"status": "success"})
	return nil
}
