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

func (m *PostgresDBRepo) OneEvent(id int) (*models.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `SELECT
							id, title, release_date,
							runtime, mpaa_rating,
							description, coalesce(image, ''),
							created_at, updated_at
						FROM
							events
						WHERE id = $1`

	row := m.DB.QueryRowContext(ctx, stmt, id)

	var movie models.Event

	err := row.Scan(
		&movie.ID,
		&movie.Title,
		&movie.ReleaseDate,
		&movie.RunTime,
		&movie.MPAARating,
		&movie.Description,
		&movie.Image,
		&movie.CreatedAt,
		&movie.UpdatedAt)

	if err != nil {
		return nil, err
	}

	stmt = `SELECT
						g.id, g.genre
					FROM
						events_genres mg
					LEFT JOIN genres g ON (mg.genre_id = g.id)
					WHERE mg.event_id = $1
					ORDER BY g.genre`

	rows, err := m.DB.QueryContext(ctx, stmt, id)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	var genres []*models.Genre
	for rows.Next() {
		var g models.Genre
		err := rows.Scan(&g.ID, &g.Genre)
		if err != nil {
			return nil, err
		}

		genres = append(genres, &g)
	}

	movie.Genres = genres

	return &movie, nil
}

func (m *PostgresDBRepo) OneEventForEdit(id int) (*models.Event, []*models.Genre, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `SELECT
						id, title, release_date, runtime,
						mpaa_rating, description, coalesce(image, ''),
						created_at, updated_at
					FROM
						events
					WHERE id = $1`

	row := m.DB.QueryRowContext(ctx, stmt, id)

	var event models.Event

	err := row.Scan(
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
		return nil, nil, err
	}

	stmt = `SELECT
						g.id, g.genre
					FROM
						movies_genres mg
					LEFT JOIN genres g ON (mg.genre_id = g.id)
					WHERE mg.movie_id = $1
					ORDER BY g.genre`

	rows, err := m.DB.QueryContext(ctx, stmt, id)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, err
	}
	defer rows.Close()

	var genres []*models.Genre
	var genresArray []int

	for rows.Next() {
		var g models.Genre
		err := rows.Scan(&g.ID, &g.Genre)
		if err != nil {
			return nil, nil, err
		}

		genres = append(genres, &g)
		genresArray = append(genresArray, g.ID)
	}

	event.Genres = genres
	event.GenresArray = genresArray

	var allGenres []*models.Genre

	stmt = `SELECT
						id, genre
					FROM
						genres
					ORDER BY genre`

	gRows, err := m.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, nil, err
	}
	defer gRows.Close()

	for gRows.Next() {
		var g models.Genre
		err := gRows.Scan(&g.ID, &g.Genre)
		if err != nil {
			return nil, nil, err
		}

		allGenres = append(allGenres, &g)
	}

	return &event, allGenres, nil
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

func (m *PostgresDBRepo) AllGenres() ([]*models.Genre, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `
		SELECT
			id, genre, created_at, updated_at
		FROM
			genres
		ORDER BY genre
		`

	rows, err := m.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []*models.Genre

	for rows.Next() {
		var g models.Genre
		err := rows.Scan(
			&g.ID,
			&g.Genre,
			&g.CreatedAt,
			&g.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		genres = append(genres, &g)
	}

	return genres, nil
}

func (m *PostgresDBRepo) InsertEvent(event models.Event) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `
		INSERT INTO events (title, description, release_date, runtime,
												mpaa_rating, created_at, updated_at, image
												VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
												RETURNING id)`

	var newID int

	err := m.DB.QueryRowContext(ctx, stmt,
		event.Title,
		event.Description,
		event.ReleaseDate,
		event.RunTime,
		event.MPAARating,
		event.CreatedAt,
		event.UpdatedAt,
		event.Image,
	).Scan(&newID)
	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (m *PostgresDBRepo) UpdateEventGenres(id int, genreIDs []int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `DELETE FROM events_genres WHERE event_id = $1`

	_, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	for _, n := range genreIDs {
		stmt = `INSERT INTO events_genres (event_id, genre_id) VALUES ($1, $2)`
		_, err := m.DB.ExecContext(ctx, stmt, id, n)
		if err != nil {
			return err
		}
	}

	return nil
}
