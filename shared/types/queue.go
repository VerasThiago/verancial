package types

import "github.com/verasthiago/verancial/shared/constants"

type ReportProcessQueuePayload struct {
	UserId string `json:"userid"`
	BankId string `json:"bankid"`

	// TODO: Design who will be responsible for knowing the filepath (will insert on db with user info?)
	FilePath string `json:"filepath"`
}

type AppIntegrationQueuePayload struct {
	UserId              string          `json:"userid"`
	AppID               constants.AppID `json:"appid"`
	BankId              string          `json:"bankid"`
	LastTransactionDate string          `json:"lasttransactiondate"`
}
