package postgresrepository

import (
	"github.com/verasthiago/verancial/shared/errors"
	"github.com/verasthiago/verancial/shared/models"
)

const USER_DATA_NAME = "user"

func (p *PostgresRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := errors.HandleDataNotFoundError(p.db.Where("email = ?", email).First(&user).Error, USER_DATA_NAME)
	return &user, err
}

func (p *PostgresRepository) CreateUser(user *models.User) error {
	return errors.HandleDuplicateError(p.db.Create(user).Error)
}

func (p *PostgresRepository) UpdateUser(user *models.User) error {
	if user.Password != "" {
		if err := user.HashPassword(user.Password); err != nil {
			return err
		}
	}
	return errors.HandleDataNotFoundError(p.db.Model(user).Updates(user).Error, USER_DATA_NAME)
}

func (p *PostgresRepository) DeleteUser(userID string) error {
	return errors.HandleDataNotFoundError(p.db.Where("id = ?", userID).Delete(&models.User{}).Error, USER_DATA_NAME)
}

func (p *PostgresRepository) MigrateUser(model *models.User) error {
	return p.db.AutoMigrate(model)
}
