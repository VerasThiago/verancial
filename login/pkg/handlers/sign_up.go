package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/verasthiago/verancial/login/pkg/builder"
	"github.com/verasthiago/verancial/login/pkg/constants"
	"github.com/verasthiago/verancial/login/pkg/validator"
	"github.com/verasthiago/verancial/shared/auth"
	"github.com/verasthiago/verancial/shared/errors"
)

type CreateUserAPI interface {
	Handler(context *gin.Context) error
}

type CreateUserHandler struct {
	builder.Builder
}

func (c *CreateUserHandler) InitFromBuilder(builder builder.Builder) *CreateUserHandler {
	c.Builder = builder
	return c
}

func (c *CreateUserHandler) Handler(context *gin.Context) error {
	var request validator.SignUpRequest
	var tokenString string
	var err error

	if err := context.ShouldBindJSON(&request); err != nil {
		return err
	}

	if errList := request.Validate(); len(errList) > 0 {
		fmt.Printf("\nerrList %+v\n", errList)
		return errors.CreateGenericErrorFromValidateError(errList)
	}

	if err := request.HashPassword(request.Password); err != nil {
		return err
	}

	if err := c.GetRepository().CreateUser(request.User); err != nil {
		return err
	}

	if tokenString, err = auth.GenerateJWT(request.User, c.GetSharedFlags().JwtKeyEmail, time.Now().Add(constants.VERIFY_EMAIL_TOKEN_EXPIRE_TIME)); err != nil {
		return err
	}

	context.JSON(http.StatusCreated, gin.H{"id": request.User.ID, "token": tokenString})
	return nil
}
