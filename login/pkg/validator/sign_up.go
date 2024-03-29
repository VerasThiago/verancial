package validator

import (
	"github.com/thedevsaddam/govalidator"
	"github.com/verasthiago/verancial/shared/models"
	"github.com/verasthiago/verancial/shared/validator"
)

type SignUpRequest struct {
	*models.User
}

func (s *SignUpRequest) Validate() []string {
	rules := govalidator.MapData{
		"name":     []string{"required"},
		"email":    []string{"required", "email"},
		"password": []string{"required"},
	}

	options := govalidator.Options{
		Data:  s,
		Rules: rules,
	}

	values := govalidator.New(options).ValidateStruct()
	return validator.MergeUrlValues([]string{"name", "email", "password"}, values)
}
