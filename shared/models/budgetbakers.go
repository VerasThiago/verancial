package models

import "time"

type BudgetBakers struct {
	Date        time.Time `json:"date"`
	Amount      float32   `json:"amount"`
	Payee       string    `json:"payee"`
	Description string    `json:"description"`
	Currency    string    `json:"currency"`
}
