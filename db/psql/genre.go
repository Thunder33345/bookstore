package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/thunder33345/bookstore"
)

// CreateGenre creates a genre using provided model
// note that ID, CreatedAt, UpdatedAt are all ignored
// returns the uuid of the created genre when successful
func (s *Store) CreateGenre(ctx context.Context, genre bookstore.Genre) (bookstore.Genre, error) {
	row := s.db.QueryRowxContext(ctx, `INSERT INTO genre(name) VALUES ($1) RETURNING *`, genre.Name)
	if err := row.Err(); err != nil {
		err = enrichPQError(err, "book.isbn")
		return bookstore.Genre{}, fmt.Errorf("creating genre.name=%s: %w", genre.Name, err)
	}

	var created bookstore.Genre
	err := row.Scan(&created)
	if err != nil {
		return bookstore.Genre{}, fmt.Errorf("scanning created genre: %w", err)
	}
	return created, nil
}

// GetGenre fetches a genre using its ID
func (s *Store) GetGenre(ctx context.Context, genreID uuid.UUID) (bookstore.Genre, error) {
	var genre bookstore.Genre
	err := s.db.GetContext(ctx, &genre, `SELECT * FROM genre WHERE id = $1 LIMIT 1`, genreID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = bookstore.NewNoResultError("genre.id", err)
		}
		return bookstore.Genre{}, fmt.Errorf("selecting genre.id=%v: %w", genreID, err)
	}
	return genre, nil
}

// ListGenres returns a list of genres
// to paginate, use the last Genre.CreatedAt you received
func (s *Store) ListGenres(ctx context.Context, limit int, after time.Time) ([]bookstore.Genre, error) {
	genres := make([]bookstore.Genre, 0, limit)
	var err error
	if !after.IsZero() {
		//after time is provided, we add WHERE created_at > after to perform pagination
		err = s.db.SelectContext(ctx, &genres, `SELECT * FROM genre WHERE created_at > $2 ORDER BY created_at LIMIT $1`, limit, after)
	} else {
		err = s.db.SelectContext(ctx, &genres, `SELECT * FROM genre ORDER BY created_at LIMIT $1`, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("listing genre with limit=%v after=%s: %w", limit, after, err)
	}
	return genres, nil
}

// UpdateGenre updates the provided genre using its ID
// note that CreatedAt, UpdatedAt cannot be set
func (s *Store) UpdateGenre(ctx context.Context, genre bookstore.Genre) error {
	if genre.ID == uuid.Nil {
		return fmt.Errorf("updating genre: %w", bookstore.ErrMissingID)
	}
	res, err := s.db.ExecContext(ctx, `UPDATE genre SET name = $1 WHERE id = $2`, genre.Name, genre.ID)
	if err != nil {
		err = enrichPQError(err, "book.isbn")
		return fmt.Errorf("updating genre: %w", err)
	}
	err = checkAffectedRows(res, bookstore.NewNoResultError("genre", err))
	if err != nil {
		return fmt.Errorf("updating genre=%v: %w", genre.ID, err)
	}
	return nil
}

// DeleteGenre deletes the specified genre using its ID
func (s *Store) DeleteGenre(ctx context.Context, genreID uuid.UUID) error {
	if genreID == uuid.Nil {
		return fmt.Errorf("missing genre id")
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM genre WHERE id = $1`, genreID)
	if err != nil {
		err = enrichPQError(err, "book.isbn")
		return fmt.Errorf("deleting genre.id=%v: %w", genreID, err)
	}
	err = checkAffectedRows(res, bookstore.NewNoResultError("genre", err))
	if err != nil {
		return fmt.Errorf("deleting genre=%v: %w", genreID, err)
	}
	return nil
}
