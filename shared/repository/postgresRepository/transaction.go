package postgresrepository

import (
	"time"

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

func (p *PostgresRepository) GetLastTransactionFromUserBank(userId string, BankId string) (*models.Transaction, error) {
	var transaction models.Transaction
	if err := p.db.Where("user_id = ? AND bank_id = ?::uuid", userId, BankId).Order("date desc").First(&transaction).Error; err != nil {
		if errors.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	transaction.Date = transaction.Date.UTC()
	return &transaction, nil
}

func (p *PostgresRepository) GetAllTransactionsFromUserBankAfterDate(userId string, BankId string, lastTransaction time.Time) ([]*models.Transaction, error) {
	var transactionList []*models.Transaction
	if err := errors.HandleDataNotFoundError(p.db.Where("user_id = ? AND bank_id = ?::uuid AND date > ? ", userId, BankId, lastTransaction).Find(&transactionList).Error, "TRANSACTIONS"); err != nil {
		return nil, err
	}

	return transactionList, nil
}

func (p *PostgresRepository) GetTransactions(userId string, bankId string, limit, offset int, filter *models.TransactionFilter) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	query := p.db.Model(&models.Transaction{}).
		Where("user_id = ? AND bank_id = ?::uuid", userId, bankId)

	if filter != nil && filter.Uncategorized {
		query = query.Where("(category IS NULL OR category = '')")
	}

	err := query.
		Order("date DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	if err != nil {
		return nil, errors.HandleDataNotFoundError(err, "TRANSACTIONS")
	}

	return transactions, nil
}

func (p *PostgresRepository) GetTransactionCountFromUserBank(userId string, BankId string) (int, error) {
	var count int64
	err := p.db.Model(&models.Transaction{}).
		Where("user_id = ? AND bank_id = ?::uuid", userId, BankId).
		Count(&count).Error
	if err != nil {
		return 0, errors.HandleDataNotFoundError(err, "TRANSACTIONS")
	}
	return int(count), nil
}
