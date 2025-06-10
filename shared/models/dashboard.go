package models

import "time"

type UserDashboardStats struct {
	TotalBankAccounts int               `json:"total_bank_accounts"`
	BankAccountStats  []BankAccountStat `json:"bank_account_stats"`
}

type BankAccountStat struct {
	BankAccount      BankAccount `json:"bank_account"`
	TransactionCount int         `json:"transaction_count"`
	LastTransaction  *time.Time  `json:"last_transaction"`
	DaysOutdated     *int        `json:"days_outdated"`
}
