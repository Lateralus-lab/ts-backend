package repository

import (
	"database/sql"

	"github.com/Lateralus-lab/ts-backend/internal/models"
)

type DatabaseRepo interface {
	Connection() *sql.DB
	AllEvents() ([]*models.Event, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id int) (*models.User, error)
}
