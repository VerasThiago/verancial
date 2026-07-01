package postgresrepository

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jackc/pgconn"
	sharederrors "github.com/verasthiago/verancial/shared/errors"
	"github.com/verasthiago/verancial/shared/models"
	"gorm.io/gorm"
)

var errDuplicateEmail = &pgconn.PgError{
	Severity: "ERROR",
	Code:     sharederrors.PSQL_DUPLICATED_KEY_ERROR_CODE,
	Message:  `duplicate key value violates unique constraint "users_email_key"`,
}

func TestPostgresRepository_GetUserByEmail(t *testing.T) {
	t.Run("returns the user on success", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		rows := sqlmock.NewRows([]string{"id", "email", "name"}).
			AddRow("user-1", "jane@example.com", "Jane")
		mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1`).
			WithArgs("jane@example.com").
			WillReturnRows(rows)

		user, err := repo.GetUserByEmail("jane@example.com")

		require.NoError(t, err)
		assert.Equal(t, "user-1", user.ID)
		assert.Equal(t, "jane@example.com", user.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns a DATA_NOT_FOUND error when no row matches", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1`).
			WithArgs("missing@example.com").
			WillReturnError(gorm.ErrRecordNotFound)

		_, err := repo.GetUserByEmail("missing@example.com")

		require.Error(t, err)
		var genericErr sharederrors.GenericError
		require.True(t, errors.As(err, &genericErr))
		assert.EqualValues(t, sharederrors.STATUS_NOT_FOUND, genericErr.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresRepository_GetUserByID(t *testing.T) {
	repo, mock := newMockRepository(t)

	rows := sqlmock.NewRows([]string{"id", "email"}).AddRow("user-1", "jane@example.com")
	mock.ExpectQuery(`SELECT \* FROM "users" WHERE id = \$1`).
		WithArgs("user-1").
		WillReturnRows(rows)

	user, err := repo.GetUserByID("user-1")

	require.NoError(t, err)
	assert.Equal(t, "user-1", user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_CreateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "users"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-1"))
		mock.ExpectCommit()

		err := repo.CreateUser(&models.User{ID: "user-1", Email: "jane@example.com"})

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate email is translated to a DATA_ALREADY_BEGIN_USED error", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "users"`).
			WillReturnError(errDuplicateEmail)
		mock.ExpectRollback()

		err := repo.CreateUser(&models.User{ID: "user-1", Email: "jane@example.com"})

		require.Error(t, err)
		var genericErr sharederrors.GenericError
		require.True(t, errors.As(err, &genericErr))
		assert.EqualValues(t, sharederrors.STATUS_BAD_REQUEST, genericErr.Code)
	})
}

func TestPostgresRepository_UpdateUser(t *testing.T) {
	t.Run("hashes a non-empty password before saving", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "users" SET`).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		user := &models.User{ID: "user-1", Password: "plaintext"}
		err := repo.UpdateUser(user)

		require.NoError(t, err)
		assert.NotEqual(t, "plaintext", user.Password, "password should have been hashed before the update was issued")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not-found is translated to a DATA_NOT_FOUND error", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "users" SET`).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectRollback()

		err := repo.UpdateUser(&models.User{ID: "missing"})

		require.Error(t, err)
		var genericErr sharederrors.GenericError
		require.True(t, errors.As(err, &genericErr))
		assert.EqualValues(t, sharederrors.STATUS_NOT_FOUND, genericErr.Code)
	})
}

func TestPostgresRepository_DeleteUser(t *testing.T) {
	// User embeds gorm.Model, so Delete is a soft delete: an UPDATE setting
	// deleted_at, not a hard DELETE statement.
	t.Run("success", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "users" SET "deleted_at"=\$1 WHERE id = \$2 AND "users"."deleted_at" IS NULL`).
			WithArgs(sqlmock.AnyArg(), "user-1").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.DeleteUser("user-1")

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not-found is translated to a DATA_NOT_FOUND error", func(t *testing.T) {
		repo, mock := newMockRepository(t)

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "users" SET "deleted_at"=\$1 WHERE id = \$2 AND "users"."deleted_at" IS NULL`).
			WithArgs(sqlmock.AnyArg(), "missing").
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectRollback()

		err := repo.DeleteUser("missing")

		require.Error(t, err)
		var genericErr sharederrors.GenericError
		require.True(t, errors.As(err, &genericErr))
		assert.EqualValues(t, sharederrors.STATUS_NOT_FOUND, genericErr.Code)
	})
}

func TestPostgresRepository_MigrateUser(t *testing.T) {
	// AutoMigrate issues a series of introspection queries against
	// information_schema; asserting a specific statement is brittle, so just
	// confirm it doesn't blow up wiring the call through to gorm and surfaces
	// a driver error when the connection is unusable.
	repo, mock := newMockRepository(t)
	mock.ExpectQuery(`SELECT`).WillReturnError(errors.New("boom"))

	err := repo.MigrateUser(&models.User{})

	assert.Error(t, err)
}
