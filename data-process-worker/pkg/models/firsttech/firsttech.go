package firsttech

import "time"

type FirstTech struct {
	TransactionID       string    `json:"transactionId"`
	PostingDate         time.Time `json:"postingDate"`
	EffectiveDate       time.Time `json:"effectiveDate"`
	TransactionType     string    `json:"transactionType"`
	Amount              float32   `json:"amount"`
	CheckNumber         string    `json:"checkNumber"`
	ReferenceNumber     string    `json:"referenceNumber"`
	Description         string    `json:"description"`
	TransactionCategory string    `json:"transactionCategory"`
	Type                string    `json:"type"`
	Balance             float32   `json:"balance"`
	Memo                string    `json:"memo"`
	ExtendedDescription string    `json:"extendedDescription"`
}
