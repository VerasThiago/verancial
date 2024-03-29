package validator

import (
	"github.com/thedevsaddam/govalidator"
	"github.com/verasthiago/verancial/shared/validator"
)

type DeleteRequest struct {
	UserId string `json:"id"`
}

func (d *DeleteRequest) Validate() []string {
	rules := govalidator.MapData{
		"id": []string{"required", "uuid"},
	}

	options := govalidator.Options{
		Data:  d,
		Rules: rules,
	}

	values := govalidator.New(options).ValidateStruct()
	return validator.MergeUrlValues([]string{"id"}, values)
}
