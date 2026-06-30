package postgresrepository

import (
	"time"

	"github.com/verasthiago/verancial/shared/errors"
	"github.com/verasthiago/verancial/shared/models"
)

func (p *PostgresRepository) MigrateBankAccount(model *models.BankAccount) error {
	return p.db.AutoMigrate(model)
}

func (p *PostgresRepository) MigrateUserBankAccount(model *models.UserBankAccount) error {
	return p.db.AutoMigrate(model)
}

func (p *PostgresRepository) GetUserBankAccounts(userId string) ([]*models.UserBankAccount, error) {
	var userBankAccounts []*models.UserBankAccount
	if err := p.db.Where("user_id = ? AND is_active = ?", userId, true).
		Preload("BankAccount").
		Find(&userBankAccounts).Error; err != nil {
		return nil, err
	}
	return userBankAccounts, nil
}

func (p *PostgresRepository) GetUserDashboardStats(userId string) (*models.UserDashboardStats, error) {
	userBankAccounts, err := p.GetUserBankAccounts(userId)
	if err != nil {
		return nil, err
	}

	stats := &models.UserDashboardStats{
		TotalBankAccounts: len(userBankAccounts),
		BankAccountStats:  make([]models.BankAccountStat, 0, len(userBankAccounts)),
	}

	if len(userBankAccounts) == 0 {
		return stats, nil
	}

	// Single aggregate query for all bank accounts instead of N+1 queries
	// (one COUNT + one "last transaction" query per bank account).
	type aggRow struct {
		BankId       string
		Count        int64
		LastTransDate *time.Time
	}
	var aggRows []aggRow
	if err := p.db.Model(&models.Transaction{}).
		Select("bank_id, COUNT(*) as count, MAX(date) as last_trans_date").
		Where("user_id = ?", userId).
		Group("bank_id").
		Scan(&aggRows).Error; err != nil {
		return nil, err
	}

	statsByBankId := make(map[string]aggRow, len(aggRows))
	for _, row := range aggRows {
		statsByBankId[row.BankId] = row
	}

	for _, userBankAccount := range userBankAccounts {
		row, found := statsByBankId[userBankAccount.BankId]

		var daysOutdated *int
		var lastTransactionTime *time.Time
		var transactionCount int

		if found {
			transactionCount = int(row.Count)
			if row.LastTransDate != nil {
				lastTransactionTime = row.LastTransDate
				days := int(time.Since(*row.LastTransDate).Hours() / 24)
				daysOutdated = &days
			}
		}

		bankAccountStat := models.BankAccountStat{
			BankAccount:      userBankAccount.BankAccount,
			TransactionCount: transactionCount,
			LastTransaction:  lastTransactionTime,
			DaysOutdated:     daysOutdated,
		}

		stats.BankAccountStats = append(stats.BankAccountStats, bankAccountStat)
	}

	return stats, nil
}

func (p *PostgresRepository) GetBankAccountById(bankId string, userId string) (*models.BankAccount, error) {
	var userBankAccount models.UserBankAccount
	err := p.db.Where("bank_id = ? AND user_id = ? AND is_active = ?", bankId, userId, true).
		Preload("BankAccount").
		First(&userBankAccount).Error
	if err != nil {
		return nil, errors.HandleDataNotFoundError(err, "BANK_ACCOUNT")
	}
	return &userBankAccount.BankAccount, nil
}
