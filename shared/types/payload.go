package types

import (
	"github.com/verasthiago/verancial/shared/constants"
)

type QueuePayload struct {
	UserID   string           `json:"userid"`
	BankName constants.BankID `json:"bankname"`

	// TODO: Design who will be responsible for knowing the filepath (will insert on db with user info?)
	FilePath string `json:"filepath"`
}
