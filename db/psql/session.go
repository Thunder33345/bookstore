package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/thunder33345/bookstore"
)

func (s *Store) StoreSession(ctx context.Context, token string, account bookstore.Account) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO session(token,account_id) VALUES ($1,$2)`, token, account.ID)
	if err != nil {
		err = enrichPQError(err, "session.token")
		return err
	}
	return nil
}

func (s *Store) GetSession(ctx context.Context, token string) (bookstore.Session, error) {
	var account bookstore.Account
	query :=
		`SELECT a.* FROM account a
			INNER JOIN session s
  			ON s.account_id = a.id
			WHERE s.token = $1;`
	err := s.db.GetContext(ctx, &account, query, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = bookstore.ErrInvalidSession
		}
		return bookstore.Session{}, fmt.Errorf("selecting session.token: %w", err)
	}
	return bookstore.Session{Account: account}, nil
}

func (s *Store) DeleteSession(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM session WHERE token = $1`, token)
	if err != nil {
		return fmt.Errorf("deleting session.token: %w", err)
	}
	return nil
}

func (s *Store) DeleteSessionsFor(ctx context.Context, accountID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM session WHERE account_id = $1`, accountID)
	if err != nil {
		return fmt.Errorf("deleting session.token: %w", err)
	}
	return nil
}
