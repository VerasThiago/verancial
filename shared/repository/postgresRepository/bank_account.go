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

	for _, userBankAccount := range userBankAccounts {
		var transactionCount int64
		var lastTransaction *models.Transaction

		if err := p.db.Model(&models.Transaction{}).
			Where("user_id = ? AND bank_id = ?", userId, userBankAccount.BankId).
			Count(&transactionCount).Error; err != nil {
			return nil, err
		}

		if err := p.db.Where("user_id = ? AND bank_id = ?", userId, userBankAccount.BankId).
			Order("date desc").
			First(&lastTransaction).Error; err != nil {
			if !errors.IsNotFoundError(err) {
				return nil, err
			}
		}

		var daysOutdated *int
		var lastTransactionTime *time.Time

		if lastTransaction != nil {
			lastTransactionTime = &lastTransaction.Date
			days := int(time.Since(lastTransaction.Date).Hours() / 24)
			daysOutdated = &days
		}

		bankAccountStat := models.BankAccountStat{
			BankAccount:      userBankAccount.BankAccount,
			TransactionCount: int(transactionCount),
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
