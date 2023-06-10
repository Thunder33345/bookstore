package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/thunder33345/bookstore"
)

func (s *Store) UpsertCoverData(ctx context.Context, cover bookstore.CoverData) (bookstore.CoverData, error) {
	query :=
		`INSERT INTO cover_data(isbn,cover_file) VALUES ($1,$2)
            ON CONFLICT DO UPDATE SET cover_file = excluded.cover_file
        RETURNING *`
	row := s.db.QueryRowxContext(ctx, query, cover.ISBN, cover.CoverFile)
	if err := row.Err(); err != nil {
		err = enrichPQError(err, "genre.name")
		return bookstore.CoverData{}, fmt.Errorf("creating cover.isbn=%s: %w", cover.ISBN, err)
	}

	var created bookstore.CoverData
	err := row.StructScan(&created)
	if err != nil {
		return bookstore.CoverData{}, fmt.Errorf("scanning created cover data: %w", err)
	}
	return created, nil
}

func (s *Store) GetCoverData(ctx context.Context, isbn string) (bookstore.CoverData, error) {
	var cover bookstore.CoverData
	err := s.db.GetContext(ctx, &cover, `SELECT * FROM cover_data WHERE isbn = $1 LIMIT 1`, isbn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = bookstore.NewNoResultError("cover.isbn", err)
		}
		return bookstore.CoverData{}, fmt.Errorf("selecting cover.isbn=%v: %w", isbn, err)
	}
	return cover, nil
}

func (s *Store) DeleteCoverData(ctx context.Context, isbn string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM cover_data WHERE isbn = $1`, isbn)
	if err != nil {
		err = enrichDeletePQError(err, "genre")
		return fmt.Errorf("deleting cover_data.isbn=%v: %w", isbn, err)
	}
	return nil
}
