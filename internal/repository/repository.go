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

	OneEventForEdit(id int) (*models.Event, []*models.Genre, error)
	OneEvent(id int) (*models.Event, error)
	AllGenres() ([]*models.Genre, error)
	InsertEvent(event models.Event) (int, error)
	UpdateEventGenres(id int, genreIDs []int) error
}
