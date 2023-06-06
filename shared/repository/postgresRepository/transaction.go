package postgresrepository

import (
	"time"

	"github.com/verasthiago/verancial/shared/constants"
	"github.com/verasthiago/verancial/shared/errors"
	"github.com/verasthiago/verancial/shared/models"
)

func (p *PostgresRepository) CreateTransactionInBatches(transactions []*models.Transaction) error {
	return errors.HandleDuplicateError(p.db.CreateInBatches(transactions, len(transactions)).Error)
}

func (p *PostgresRepository) MigrateTransaction(model *models.Transaction) error {
	return p.db.AutoMigrate(model)
}

func (p *PostgresRepository) GetAllTransactionsFromUserBankAfterDate(userId string, BankId constants.BankId, lastTransaction time.Time) ([]*models.Transaction, error) {
	var transactionList []*models.Transaction
	if err := errors.HandleDataNotFoundError(p.db.Where("user_id = ? AND bank_id = ? and date > ? ", userId, BankId, lastTransaction).Find(&transactionList).Error, "TRANSACTIONS"); err != nil {
		return nil, err
	}

	return transactionList, nil
}
