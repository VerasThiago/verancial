package models

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
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
	BankId      string              `json:"bank_id" gorm:"type:uuid"`
	Metadata    TransactionMetadata `json:"metadata" gorm:"type:jsonb"`
	Fingerprint string              `json:"fingerprint" gorm:"uniqueIndex:idx_transaction_fingerprint"`
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

func (t *Transaction) SetFingerprint() {
	normalizedDate := t.Date.UTC().Format("2006-01-02")
	normalizedAmount := fmt.Sprintf("%.2f", t.Amount)
	normalizedDescription := strings.ToLower(strings.TrimSpace(t.Description))

	fingerprintData := fmt.Sprintf("%s|%s|%s|%s",
		t.UserId,
		normalizedDate,
		normalizedAmount,
		normalizedDescription,
	)

	hash := sha256.Sum256([]byte(fingerprintData))
	t.Fingerprint = fmt.Sprintf("%x", hash)
}

type TransactionFilter struct {
	Uncategorized bool
}
