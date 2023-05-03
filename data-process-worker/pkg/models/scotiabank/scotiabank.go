package scotiabank

import "time"

type ScotiaBank struct {
	Date        time.Time `json:"date"`
	Amount      float32   `json:"amount"`
	Description string    `json:"description"`
	Payee       string    `json:"payee"`
}
