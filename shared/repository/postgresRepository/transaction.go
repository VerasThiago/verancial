package postgresrepository

import (
	"github.com/verasthiago/verancial/shared/errors"
	"github.com/verasthiago/verancial/shared/models"
)

func (p *PostgresRepository) CreateTransactionInBatches(transactions []*models.Transaction) error {
	return errors.HandleDuplicateError(p.db.CreateInBatches(transactions, len(transactions)).Error)
}

func (p *PostgresRepository) MigrateTransaction(model *models.Transaction) error {
	return p.db.AutoMigrate(model)
}
