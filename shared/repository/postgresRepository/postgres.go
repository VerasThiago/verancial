package postgresrepository

import (
	"fmt"

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

	fmt.Println("Connected to Database!")

	return &PostgresRepository{
		db,
	}
}
