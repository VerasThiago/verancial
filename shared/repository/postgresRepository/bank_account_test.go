package postgresrepository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sharederrors "github.com/verasthiago/verancial/shared/errors"
	"github.com/verasthiago/verancial/shared/models"
	"gorm.io/gorm"
)

func TestPostgresRepository_GetUserBankAccounts(t *testing.T) {
	t.Run("returns active bank accounts with the BankAccount relation preloaded", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectQuery(`SELECT \* FROM "user_bank_accounts" WHERE user_id = \$1 AND is_active = \$2`).
			WithArgs("user-1", true).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "bank_id", "is_active"}).
				AddRow("uba-1", "user-1", "bank-1", true))
		mock.ExpectQuery(`SELECT \* FROM "bank_accounts" WHERE "bank_accounts"\."id" = \$1`).
			WithArgs("bank-1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("bank-1", "Scotiabank"))

		accounts, err := repo.GetUserBankAccounts("user-1")

		require.NoError(t, err)
		require.Len(t, accounts, 1)
		assert.Equal(t, "uba-1", accounts[0].ID)
		assert.Equal(t, "Scotiabank", accounts[0].BankAccount.Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("propagates query errors", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectQuery(`SELECT \* FROM "user_bank_accounts"`).
			WillReturnError(errBoom)

		_, err := repo.GetUserBankAccounts("user-1")

		assert.Error(t, err)
	})
}

func TestPostgresRepository_GetBankAccountById(t *testing.T) {
	t.Run("returns the bank account when found", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectQuery(`SELECT \* FROM "user_bank_accounts" WHERE bank_id = \$1 AND user_id = \$2 AND is_active = \$3`).
			WithArgs("bank-1", "user-1", true).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "bank_id", "is_active"}).
				AddRow("uba-1", "user-1", "bank-1", true))
		mock.ExpectQuery(`SELECT \* FROM "bank_accounts" WHERE "bank_accounts"\."id" = \$1`).
			WithArgs("bank-1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("bank-1", "Scotiabank"))

		account, err := repo.GetBankAccountById("bank-1", "user-1")

		require.NoError(t, err)
		assert.Equal(t, "bank-1", account.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not-found is translated to a DATA_NOT_FOUND error", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectQuery(`SELECT \* FROM "user_bank_accounts" WHERE bank_id = \$1 AND user_id = \$2 AND is_active = \$3`).
			WithArgs("bank-1", "user-1", true).
			WillReturnError(gorm.ErrRecordNotFound)

		_, err := repo.GetBankAccountById("bank-1", "user-1")

		require.Error(t, err)
		var genericErr sharederrors.GenericError
		require.ErrorAs(t, err, &genericErr)
		assert.EqualValues(t, sharederrors.STATUS_NOT_FOUND, genericErr.Code)
	})
}

func TestPostgresRepository_GetUserDashboardStats(t *testing.T) {
	t.Run("aggregates transaction counts/last-transaction-date per bank account in a single query", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectQuery(`SELECT \* FROM "user_bank_accounts" WHERE user_id = \$1 AND is_active = \$2`).
			WithArgs("user-1", true).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "bank_id", "is_active"}).
				AddRow("uba-1", "user-1", "bank-1", true))
		mock.ExpectQuery(`SELECT \* FROM "bank_accounts" WHERE "bank_accounts"\."id" = \$1`).
			WithArgs("bank-1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("bank-1", "Scotiabank"))

		mock.ExpectQuery(`SELECT bank_id, COUNT\(\*\) as count, MAX\(date\) as last_trans_date FROM "transactions" WHERE user_id = \$1`).
			WithArgs("user-1").
			WillReturnRows(sqlmock.NewRows([]string{"bank_id", "count", "last_trans_date"}).
				AddRow("bank-1", int64(3), nil))

		stats, err := repo.GetUserDashboardStats("user-1")

		require.NoError(t, err)
		assert.Equal(t, 1, stats.TotalBankAccounts)
		require.Len(t, stats.BankAccountStats, 1)
		assert.Equal(t, 3, stats.BankAccountStats[0].TransactionCount)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no bank accounts short-circuits without an aggregate query", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectQuery(`SELECT \* FROM "user_bank_accounts" WHERE user_id = \$1 AND is_active = \$2`).
			WithArgs("user-1", true).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "bank_id", "is_active"}))

		stats, err := repo.GetUserDashboardStats("user-1")

		require.NoError(t, err)
		assert.Equal(t, 0, stats.TotalBankAccounts)
		assert.Empty(t, stats.BankAccountStats)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("propagates GetUserBankAccounts errors", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectQuery(`SELECT \* FROM "user_bank_accounts"`).
			WillReturnError(errBoom)

		_, err := repo.GetUserDashboardStats("user-1")

		assert.Error(t, err)
	})
}

func TestPostgresRepository_MigrateBankAccount(t *testing.T) {
	repo, mock := newMockRepository(t)
	mock.ExpectQuery(`SELECT`).WillReturnError(errBoom)

	err := repo.MigrateBankAccount(&models.BankAccount{})

	assert.Error(t, err)
}

func TestPostgresRepository_MigrateUserBankAccount(t *testing.T) {
	repo, mock := newMockRepository(t)
	mock.ExpectQuery(`SELECT`).WillReturnError(errBoom)

	err := repo.MigrateUserBankAccount(&models.UserBankAccount{})

	assert.Error(t, err)
}
