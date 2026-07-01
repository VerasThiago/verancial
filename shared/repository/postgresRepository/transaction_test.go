package postgresrepository

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/verasthiago/verancial/shared/models"
	"gorm.io/gorm"
)

var errBoom = errors.New("boom")

func TestPostgresRepository_CreateTransactionInBatches(t *testing.T) {
	t.Run("empty slice still begins/commits an empty batch", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectBegin()
		mock.ExpectCommit()

		err := repo.CreateTransactionInBatches(nil)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("inserts transactions", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "transactions"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("tx-1"))
		mock.ExpectCommit()

		err := repo.CreateTransactionInBatches([]*models.Transaction{
			{ID: "tx-1", UserId: "user-1", Amount: 10, Date: time.Now()},
		})

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate key is translated to a DATA_ALREADY_BEGIN_USED error", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "transactions"`).
			WillReturnError(errDuplicateEmail)
		mock.ExpectRollback()

		err := repo.CreateTransactionInBatches([]*models.Transaction{
			{ID: "tx-1", UserId: "user-1", Amount: 10, Date: time.Now()},
		})

		require.Error(t, err)
	})
}

func TestPostgresRepository_MigrateTransaction(t *testing.T) {
	repo, mock := newMockRepository(t)
	mock.ExpectQuery(`SELECT`).WillReturnError(errBoom)

	err := repo.MigrateTransaction(&models.Transaction{})

	assert.Error(t, err)
}

func TestPostgresRepository_CreateUniqueTransactionInBatches(t *testing.T) {
	t.Run("empty slice is a no-op (no DB call at all)", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		err := repo.CreateUniqueTransactionInBatches(nil)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("inserts with ON CONFLICT DO NOTHING on the fingerprint column", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "transactions".*ON CONFLICT \("fingerprint"\) DO NOTHING`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("tx-1"))
		mock.ExpectCommit()

		err := repo.CreateUniqueTransactionInBatches([]*models.Transaction{
			{ID: "tx-1", UserId: "user-1", Amount: 10, Date: time.Now(), Fingerprint: "abc"},
		})

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_GetLastTransactionFromUserBank(t *testing.T) {
	t.Run("returns the most recent transaction", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		date := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		rows := sqlmock.NewRows([]string{"id", "user_id", "bank_id", "date"}).
			AddRow("tx-1", "user-1", "bank-1", date)
		mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE \(user_id = \$1 AND bank_id = \$2::uuid\) AND "transactions"\."deleted_at" IS NULL ORDER BY date desc`).
			WithArgs("user-1", "bank-1").
			WillReturnRows(rows)

		tx, err := repo.GetLastTransactionFromUserBank("user-1", "bank-1")

		require.NoError(t, err)
		require.NotNil(t, tx)
		assert.Equal(t, "tx-1", tx.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns nil, nil when no transaction exists (not an error)", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE \(user_id = \$1 AND bank_id = \$2::uuid\)`).
			WithArgs("user-1", "bank-1").
			WillReturnError(gorm.ErrRecordNotFound)

		tx, err := repo.GetLastTransactionFromUserBank("user-1", "bank-1")

		require.NoError(t, err)
		assert.Nil(t, tx)
	})

	t.Run("propagates other errors", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE \(user_id = \$1 AND bank_id = \$2::uuid\)`).
			WithArgs("user-1", "bank-1").
			WillReturnError(errBoom)

		_, err := repo.GetLastTransactionFromUserBank("user-1", "bank-1")

		assert.Error(t, err)
	})
}

func TestPostgresRepository_GetAllTransactionsFromUserBankAfterDate(t *testing.T) {
	repo, mock := newMockRepository(t)

	after := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "user_id", "bank_id", "date"}).
		AddRow("tx-1", "user-1", "bank-1", after.Add(24*time.Hour)).
		AddRow("tx-2", "user-1", "bank-1", after.Add(48*time.Hour))

	mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE \(user_id = \$1 AND bank_id = \$2::uuid AND date > \$3`).
		WithArgs("user-1", "bank-1", after).
		WillReturnRows(rows)

	transactions, err := repo.GetAllTransactionsFromUserBankAfterDate("user-1", "bank-1", after)

	require.NoError(t, err)
	assert.Len(t, transactions, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_GetTransactions(t *testing.T) {
	t.Run("without filter", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		rows := sqlmock.NewRows([]string{"id"}).AddRow("tx-1")
		mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE \(user_id = \$1 AND bank_id = \$2::uuid\) AND "transactions"\."deleted_at" IS NULL ORDER BY date DESC LIMIT 10`).
			WithArgs("user-1", "bank-1").
			WillReturnRows(rows)

		transactions, err := repo.GetTransactions("user-1", "bank-1", 10, 0, nil)

		require.NoError(t, err)
		assert.Len(t, transactions, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("with Uncategorized filter adds a category clause", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		rows := sqlmock.NewRows([]string{"id"}).AddRow("tx-1")
		mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE \(user_id = \$1 AND bank_id = \$2::uuid\) AND \(\(category IS NULL OR category = ''\)\)`).
			WithArgs("user-1", "bank-1").
			WillReturnRows(rows)

		transactions, err := repo.GetTransactions("user-1", "bank-1", 10, 0, &models.TransactionFilter{Uncategorized: true})

		require.NoError(t, err)
		assert.Len(t, transactions, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("propagates query errors", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectQuery(`SELECT \* FROM "transactions"`).
			WillReturnError(errBoom)

		_, err := repo.GetTransactions("user-1", "bank-1", 10, 0, nil)

		assert.Error(t, err)
	})
}

func TestPostgresRepository_GetTransactionCountFromUserBank(t *testing.T) {
	repo, mock := newMockRepository(t)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(5)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "transactions" WHERE \(user_id = \$1 AND bank_id = \$2::uuid\)`).
		WithArgs("user-1", "bank-1").
		WillReturnRows(rows)

	count, err := repo.GetTransactionCountFromUserBank("user-1", "bank-1")

	require.NoError(t, err)
	assert.Equal(t, 5, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}
