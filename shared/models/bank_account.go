package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BankAccount struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	DisplayName string    `json:"display_name" db:"display_name"`
	CountryCode string    `json:"country_code" db:"country_code"`
	Currency    string    `json:"currency" db:"currency"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsActive    bool      `json:"is_active" db:"is_active"`
}

func (b *BankAccount) TableName() string {
	return "bank_accounts"
}

type UserBankAccount struct {
	ID           string     `json:"id" gorm:"primary_key"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	UserId       string     `json:"user_id"`
	BankId       string     `json:"bank_id"`
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	LastSyncDate *time.Time `json:"last_sync_date"`

	// Relationships
	User        User        `json:"user,omitempty" gorm:"foreignKey:UserId"`
	BankAccount BankAccount `json:"bank_account,omitempty" gorm:"foreignKey:BankId"`
}

func (uba *UserBankAccount) TableName() string {
	return "user_bank_accounts"
}

func (uba *UserBankAccount) BeforeCreate(tx *gorm.DB) (err error) {
	uba.ID = uuid.New().String()
	return nil
}
