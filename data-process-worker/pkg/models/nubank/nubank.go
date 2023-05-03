package nubank

import "time"

type Nubank struct {
	Date          time.Time `json:"date"`
	Amount        float32   `json:"amount"`
	TransactionID string    `json:"transactionId"`
	Description   string    `json:"description"`
	Payee         string    `json:"payee"`
}
