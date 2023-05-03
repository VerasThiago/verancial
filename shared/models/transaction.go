package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Transaction struct {
	gorm.Model
	ID          string            `json:"id" gorm:"primary_key"`
	Date        time.Time         `json:"date"`
	Amount      float32           `json:"amount"`
	Payee       string            `json:"payee"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Metadata    map[string]string `json:"metadata" gorm:"type:jsonb"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New().String()
	return nil
}
