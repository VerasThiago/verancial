package repository

import (
	"github.com/verasthiago/verancial/shared/models"
)

type Repository interface {
	GetUserByEmail(email string) (*models.User, error)
	CreateUser(user *models.User) error
	DeleteUser(userID string) error
	UpdateUser(user *models.User) error

	MigrateUser(model *models.User) error
}