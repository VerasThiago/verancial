package repository

import (
	"time"

	"github.com/verasthiago/verancial/shared/constants"
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
	GetAllTransactionsFromUserBankAfterDate(userId string, BankId constants.BankId, lastTransaction time.Time) ([]*models.Transaction, error)
	GetLastTransactionFromUserBank(userId string, BankId constants.BankId) (*models.Transaction, error)

	MigrateUser(model *models.User) error
	MigrateTransaction(model *models.Transaction) error
}
