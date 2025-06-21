package repository

import (
	"time"

	"github.com/verasthiago/verancial/shared/models"
)

type Repository interface {
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id string) (*models.User, error)
	CreateUser(user *models.User) error
	DeleteUser(userId string) error
	UpdateUser(user *models.User) error

	CreateTransactionInBatches(transacions []*models.Transaction) error
	CreateUniqueTransactionInBatches(transactions []*models.Transaction) error
	GetAllTransactionsFromUserBankAfterDate(userId string, BankId string, lastTransaction time.Time) ([]*models.Transaction, error)
	GetLastTransactionFromUserBank(userId string, BankId string) (*models.Transaction, error)
	GetTransactionCountFromUserBank(userId string, BankId string) (int, error)

	MigrateUser(model *models.User) error
	MigrateTransaction(model *models.Transaction) error
	MigrateUserBankAccount(model *models.UserBankAccount) error
	MigrateBankAccount(model *models.BankAccount) error

	GetUserBankAccounts(userId string) ([]*models.UserBankAccount, error)
	GetUserDashboardStats(userId string) (*models.UserDashboardStats, error)
	GetBankAccountById(bankId string, userId string) (*models.BankAccount, error)

	GetTransactions(userId string, bankId string, limit, offset int, filter *models.TransactionFilter) ([]*models.Transaction, error)
}
