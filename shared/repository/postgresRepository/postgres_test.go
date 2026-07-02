package postgresrepository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// newMockRepository opens a gorm DB backed by a sqlmock connection so
// PostgresRepository methods can be tested against asserted SQL/args and
// canned result rows without a real Postgres instance.
func newMockRepository(t *testing.T) (*PostgresRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	return &PostgresRepository{db: gormDB}, mock
}
