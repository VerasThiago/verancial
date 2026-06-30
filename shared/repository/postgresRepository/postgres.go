package postgresrepository

import (
	"fmt"
	"time"

	shared "github.com/verasthiago/verancial/shared/flags"
	"github.com/verasthiago/verancial/shared/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresRepository struct {
	db *gorm.DB
}

func (p *PostgresRepository) InitFromFlags(sharedFlags *shared.SharedFlags) repository.Repository {
	var dsn string

	if sharedFlags.DatabaseURL != "" {
		dsn = sharedFlags.DatabaseURL
	} else {
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
			sharedFlags.DatabaseHost,
			sharedFlags.DatabaseUser,
			sharedFlags.DatabasePassword,
			sharedFlags.DatabaseName,
			sharedFlags.DatabasePort,
			sharedFlags.DatabaseSSLMode,
			sharedFlags.DatabaseTimeZone)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Cannot connect to DB")
	}

	// Configure connection pooling so the service doesn't open an unbounded
	// number of Postgres connections under load (and idle connections get
	// recycled instead of being held open indefinitely).
	sqlDB, err := db.DB()
	if err != nil {
		panic("Cannot get underlying sql.DB from gorm")
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	fmt.Println("Connected to Database!")

	return &PostgresRepository{
		db,
	}
}
