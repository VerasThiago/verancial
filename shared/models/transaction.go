package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/verasthiago/verancial/shared/constants"
	"gorm.io/gorm"
)

type Transaction struct {
	gorm.Model
	ID          string              `json:"id" gorm:"primary_key"`
	UserId      string              `json:"userid"`
	Date        time.Time           `json:"date"`
	Amount      float32             `json:"amount"`
	Payee       string              `json:"payee"`
	Description string              `json:"description"`
	Category    string              `json:"category"`
	Currency    string              `json:"currency"`
	BankId      constants.BankId    `json:"bankid"`
	Metadata    TransactionMetadata `json:"metadata" gorm:"type:jsonb"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New().String()
	return nil
}

type TransactionMetadata map[string]string

func (t *TransactionMetadata) Scan(value interface{}) error {
	if value == nil {
		*t = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("TransactionMetadata.Scan: unsupported value type %T", value)
	}
	return json.Unmarshal(b, t)
}
