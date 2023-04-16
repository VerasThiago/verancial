package types

import "github.com/verasthiago/verancial/shared/constants"

type QueuePayload struct {
	FilePath string             `json:"filepath"`
	BankName constants.BankName `json:"bankname"`
}
