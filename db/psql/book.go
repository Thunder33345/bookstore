package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nullism/bqb"
	"github.com/thunder33345/bookstore"
)

// CreateBook creates a book using provided model
// note that CreatedAt, UpdatedAt are ignored
// returns the uuid of the created book when successful
func (s *Store) CreateBook(ctx context.Context, book bookstore.Book) (bookstore.Book, error) {
	row := s.db.QueryRowxContext(ctx,
		`INSERT INTO book(isbn,title,publish_year,fiction,author_id,genre_id)
				VALUES ($1,$2,$3,$4,$5,$6) RETURNING *`, book.ISBN, book.Title, book.PublishYear, book.Fiction, book.AuthorID, book.GenreID)
	if err := row.Err(); err != nil {
		err = enrichPQError(err, "book.isbn")
		return bookstore.Book{}, fmt.Errorf("creating book: %w", err)
	}

	var created bookstore.Book
	err := row.StructScan(&created)
	if err != nil {
		return bookstore.Book{}, fmt.Errorf("scanning created book: %w", err)
	}
	return created, nil
}

// GetBook fetches n book using its ID
func (s *Store) GetBook(ctx context.Context, bookID string) (bookstore.Book, error) {
	var book bookstore.Book
	err := s.db.GetContext(ctx, &book, `SELECT b.*, c.cover_file FROM book b LEFT JOIN cover_data c ON b.isbn = c.isbn WHERE b.isbn = $1 LIMIT 1`, bookID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = bookstore.NewNoResultError("book.isbn", err)
		}
		return bookstore.Book{}, fmt.Errorf("selecting book.isbn=%v: %w", bookID, err)
	}
	return book, nil
}

// ListBooks returns a list of books
// to paginate, use the last Book.ISBN you received
// you can filter using a list of genre and author ids
// it will return if a book matches one of the provided authors and genres
// leaving it blank will omit filtering
// searchTitle performs fuzzy searching on the title of the book
func (s *Store) ListBooks(ctx context.Context, limit int, after string, genresId []uuid.UUID, authorsId []uuid.UUID, searchTitle string) ([]bookstore.Book, error) {
	books := make([]bookstore.Book, 0, limit)
	var err error
	//using bqb to build more complicated queries
	//sel is on the beginning simply for readability
	sel := bqb.New(`SELECT b.*, c.cover_file FROM book b LEFT JOIN cover_data c ON b.isbn = c.isbn`)

	where := bqb.Optional(`WHERE`)
	if after != "" {
		//if after uuid is provided, we add WHERE created_at > after via sub query to perform pagination
		//we use COALESCE to trigger a function that raises error if the selected ISBN does not exist
		where.Space(`b.created_at > COALESCE((SELECT created_at FROM book WHERE isbn = ?),raise_error_tz('Nonexistent ISBN'))`, after)
	}
	if len(genresId) > 0 {
		where.And(`b.genre_id IN (?)`, genresId)
	}
	if len(authorsId) > 0 {
		where.And(`b.author_id IN (?)`, authorsId)
	}

	//we set the order to allow overwriting it when searching
	order := bqb.New(`ORDER BY created_at`)
	if searchTitle != "" {
		where.And(`SIMILARITY(b.title, ?) > 0.1`, searchTitle)
		order = bqb.New(`ORDER BY SIMILARITY(b.title, ?) DESC`, searchTitle)
	}
	q := bqb.New(`? ? ? LIMIT ?`, sel, where, order, limit)

	query, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("bqb building query: %w", err)
	}

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return nil, fmt.Errorf("sqlx building query: %w", err)
	}
	query = s.db.Rebind(query)
	err = s.db.SelectContext(ctx, &books, query, args...)

	err = enrichListPQError(err, "book")
	if err != nil {
		return nil, fmt.Errorf("selecting book limit=%v after=%s genres=%v authors=%v: %w", limit, after, genresId, authorsId, err)
	}
	return books, nil
}

// UpdateBook updates the provided book using its ID
// note that CreatedAt, UpdatedAt cannot be set
func (s *Store) UpdateBook(ctx context.Context, book bookstore.Book) error {
	if book.ISBN == "" {
		return bookstore.ErrMissingID
	}

	//we use query builder to create optional updates book dates
	opt := bqb.Optional("")
	if !book.CreatedAt.IsZero() {
		opt.Comma(`created_at = $1`, book.CreatedAt)
	}
	if !book.UpdatedAt.IsZero() {
		opt.Comma(`updated_at = $1`, book.UpdatedAt)
	}
	q := bqb.New(`UPDATE book SET title = ?, publish_year = ?, fiction = ?, author_id = ?, genre_id = ? ? WHERE isbn = ?`,
		book.Title, book.PublishYear, book.Fiction, book.AuthorID, book.GenreID, opt, book.ISBN)
	query, args, err := q.ToPgsql()
	if err != nil {
		return fmt.Errorf("bqb building query: %w", err)
	}

	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		err = enrichPQError(err, "book.isbn")
		return fmt.Errorf("error updating book: %w", err)
	}
	err = checkAffectedRows(res, bookstore.NewNoResultError("book", err))
	if err != nil {
		return fmt.Errorf("updating book=%s: %w", book.ISBN, err)
	}

	return nil
}

// DeleteBook deletes the specified book using its ID
func (s *Store) DeleteBook(ctx context.Context, bookID string) error {
	if bookID == "" {
		return fmt.Errorf("missing book id")
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM book WHERE isbn = $1`, bookID)
	if err != nil {
		return fmt.Errorf("deleting book=%v: %w", bookID, err)
	}
	err = checkAffectedRows(res, bookstore.NewNoResultError("book", err))
	if err != nil {
		return fmt.Errorf("deleting book=%v: %w", bookID, err)
	}
	return nil
}
