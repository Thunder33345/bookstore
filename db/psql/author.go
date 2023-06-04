package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/thunder33345/bookstore"
)

// CreateAuthor creates an author using provided model
// note that ID, CreatedAt, UpdatedAt are all ignored
// returns the uuid of the created author when successful
func (s *Store) CreateAuthor(ctx context.Context, author bookstore.Author) (bookstore.Author, error) {
	row := s.db.QueryRowxContext(ctx, `INSERT INTO author(name) VALUES ($1) RETURNING *`, author.Name)
	if err := row.Err(); err != nil {
		err = enrichPQError(err, "author.name")
		return bookstore.Author{}, fmt.Errorf("creating author.name=%s: %w", author.Name, err)
	}

	var created bookstore.Author
	err := row.StructScan(&created)
	if err != nil {
		return bookstore.Author{}, fmt.Errorf("scanning created author: %w", err)
	}
	return created, nil
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
func (s *Store) ListAuthors(ctx context.Context, limit int, after uuid.UUID) ([]bookstore.Author, error) {
	authors := make([]bookstore.Author, 0, limit)
	var err error
	if after != uuid.Nil {
		//if after uuid is provided, we add WHERE created_at > after via sub query to perform pagination
		//we use COALESCE to trigger a function that raises error if the selected ID does not exist
		query := `SELECT * FROM author WHERE created_at > COALESCE((SELECT created_at FROM author WHERE id = $2),raise_error_tz('Nonexistent UUID')) ORDER BY created_at LIMIT $1`
		err = s.db.SelectContext(ctx, &authors, query, limit, after)
	} else {
		err = s.db.SelectContext(ctx, &authors, `SELECT * FROM author ORDER BY created_at LIMIT $1`, limit)
	}
	err = enrichListPQError(err, after, "author")

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
		err = enrichDeletePQError(err, "author")
		return fmt.Errorf("deleting author.id=%v: %w", authorID, err)
	}
	err = checkAffectedRows(res, bookstore.NewNoResultError("author", err))
	if err != nil {
		return fmt.Errorf("deleting author=%v: %w", authorID, err)
	}
	return nil
}
