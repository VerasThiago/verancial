package postgresrepository

import (
	"time"

	"github.com/verasthiago/verancial/shared/constants"
	"github.com/verasthiago/verancial/shared/errors"
	"github.com/verasthiago/verancial/shared/models"
	"gorm.io/gorm/clause"
)

func (p *PostgresRepository) CreateTransactionInBatches(transactions []*models.Transaction) error {
	return errors.HandleDuplicateError(p.db.CreateInBatches(transactions, len(transactions)).Error)
}

func (p *PostgresRepository) CreateUniqueTransactionInBatches(transactions []*models.Transaction) error {
	if len(transactions) == 0 {
		return nil
	}

	return errors.HandleDuplicateError(p.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "fingerprint"},
		},
		DoNothing: true,
	}).CreateInBatches(transactions, len(transactions)).Error)
}

func (p *PostgresRepository) MigrateTransaction(model *models.Transaction) error {
	return p.db.AutoMigrate(model)
}

func (p *PostgresRepository) GetLastTransactionFromUserBank(userId string, BankId constants.BankId) (*models.Transaction, error) {
	var transaction models.Transaction
	if err := p.db.Where("user_id = ? AND bank_id = ?", userId, BankId).Order("date desc").First(&transaction).Error; err != nil {
		if errors.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	transaction.Date = transaction.Date.UTC()
	return &transaction, nil
}

func (p *PostgresRepository) GetAllTransactionsFromUserBankAfterDate(userId string, BankId constants.BankId, lastTransaction time.Time) ([]*models.Transaction, error) {
	var transactionList []*models.Transaction
	if err := errors.HandleDataNotFoundError(p.db.Where("user_id = ? AND bank_id = ? and date > ? ", userId, BankId, lastTransaction).Find(&transactionList).Error, "TRANSACTIONS"); err != nil {
		return nil, err
	}

	return transactionList, nil
}
