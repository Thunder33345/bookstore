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

// CreateAuthor creates an author using provided model
// note that ID, CreatedAt, UpdatedAt are all ignored
// returns the uuid of the created author when successful
func (s *Store) CreateAuthor(ctx context.Context, author bookstore.Author) (uuid.UUID, error) {
	row := s.db.QueryRowxContext(ctx, `INSERT INTO author(name) VALUES ($1) RETURNING id`, author.Name)
	if err := row.Err(); err != nil {
		err = enrichPQError(err, "author.name")
		return uuid.Nil, fmt.Errorf("creating author.name=%s: %w", author.Name, err)
	}
	var id uuid.UUID
	err := row.Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("scanning created author id: %w", err)
	}
	return id, nil
}

// GetAuthor fetches an author using its ID
func (s *Store) GetAuthor(ctx context.Context, authorID uuid.UUID) (bookstore.Author, error) {
	var author bookstore.Author
	err := s.db.GetContext(ctx, &author, `SELECT * FROM author WHERE id = $1 LIMIT 1`, authorID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = bookstore.NewNoResultError("author.id", err)
		}
		return bookstore.Author{}, fmt.Errorf("selecting author.id=%v: %w", authorID, err)
	}
	return author, nil
}

// ListAuthors returns a list of authors
// to paginate, use the last Author.CreatedAt you received
func (s *Store) ListAuthors(ctx context.Context, limit int, after time.Time) ([]bookstore.Author, error) {
	authors := make([]bookstore.Author, 0, limit)
	var err error
	if !after.IsZero() {
		//after time is provided, we add WHERE created_at > after to perform pagination
		err = s.db.SelectContext(ctx, &authors, `SELECT * FROM author WHERE created_at > $2 ORDER BY created_at LIMIT $1`, limit, after)
	} else {
		err = s.db.SelectContext(ctx, &authors, `SELECT * FROM author ORDER BY created_at LIMIT $1`, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("listing authors limit=%v after=%s: %w", limit, after, err)
	}
	return authors, nil
}

// UpdateAuthor updates the provided author using its ID
// note that CreatedAt, UpdatedAt cannot be set
func (s *Store) UpdateAuthor(ctx context.Context, author bookstore.Author) error {
	if author.ID == uuid.Nil {
		return bookstore.ErrMissingID
	}
	res, err := s.db.ExecContext(ctx, `UPDATE author SET name = $1 WHERE id = $2`, author.Name, author.ID)
	if err != nil {
		err = enrichPQError(err, "author.name")
		return fmt.Errorf("updating author: %w", err)
	}
	err = checkAffectedRows(res, bookstore.NewNoResultError("author", err))
	if err != nil {
		return fmt.Errorf("updating author=%v: %w", author.ID, err)
	}
	return nil
}

// DeleteAuthor deletes the specified author using its ID
func (s *Store) DeleteAuthor(ctx context.Context, authorID uuid.UUID) error {
	if authorID == uuid.Nil {
		return fmt.Errorf("missing author id")
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM author WHERE id = $1`, authorID)
	if err != nil {
		err = enrichPQError(err, "author")
		return fmt.Errorf("deleting author.id=%v: %w", authorID, err)
	}
	err = checkAffectedRows(res, bookstore.NewNoResultError("author", err))
	if err != nil {
		return fmt.Errorf("deleting author=%v: %w", authorID, err)
	}
	return nil
}
