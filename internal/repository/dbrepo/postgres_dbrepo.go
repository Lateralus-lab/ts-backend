package dbrepo

import (
	"context"
	"database/sql"
	"time"

	"github.com/Lateralus-lab/ts-backend/internal/models"
)

type PostgresDBRepo struct {
	DB *sql.DB
}

const dbTimeout = time.Second * 5

func (m *PostgresDBRepo) Connection() *sql.DB {
	return m.DB
}

func (m *PostgresDBRepo) AllEvents() ([]*models.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `
			SELECT 
				id, title, release_date, runtime,
				mpaa_rating, description, coalesce(image, ''),
				created_at, updated_at
			FROM 
				events
			ORDER BY 
				title
			`

	rows, err := m.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.Event

	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.ReleaseDate,
			&event.RunTime,
			&event.MPAARating,
			&event.Description,
			&event.Image,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		events = append(events, &event)
	}

	return events, nil
}

func (m *PostgresDBRepo) GetUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `
		SELECT 
			id, email, first_name, last_name, password, created_at, updated_at
		FROM 
			users
		WHERE 
			email = $1
	`

	var user models.User
	row := m.DB.QueryRowContext(ctx, stmt, email)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (m *PostgresDBRepo) GetUserByID(id int) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `
		SELECT 
			id, email, first_name, last_name, password, created_at, updated_at
		FROM 
			users
		WHERE 
			id = $1
	`

	var user models.User
	row := m.DB.QueryRowContext(ctx, stmt, id)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
