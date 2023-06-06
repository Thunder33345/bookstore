package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/thunder33345/bookstore"
)

// CreateAccount creates an account using provided model
// note that ID, CreatedAt, UpdatedAt are all ignored
// returns the created account when successful
func (s *Store) CreateAccount(ctx context.Context, account bookstore.Account) (bookstore.Account, error) {
	row := s.db.QueryRowxContext(ctx, `INSERT INTO account(name,email,password_hash,is_admin) VALUES ($1,$2,$3,$4) RETURNING *`,
		account.Name, account.Email, account.PasswordHash, account.Admin)
	if err := row.Err(); err != nil {
		err = enrichPQError(err, "account.email")
		return bookstore.Account{}, fmt.Errorf("creating account.name=%s: %w", account.Name, err)
	}

	var created bookstore.Account
	err := row.StructScan(&created)
	if err != nil {
		return bookstore.Account{}, fmt.Errorf("scanning created account: %w", err)
	}
	return created, nil
}

// GetAccount fetches an account using its ID
func (s *Store) GetAccount(ctx context.Context, accountID uuid.UUID) (bookstore.Account, error) {
	var account bookstore.Account
	err := s.db.GetContext(ctx, &account, `SELECT * FROM account WHERE id = $1 LIMIT 1`, accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = bookstore.NewNoResultError("account.id", err)
		}
		return bookstore.Account{}, fmt.Errorf("selecting account.id=%v: %w", accountID, err)
	}
	return account, nil
}

// GetAccountByEmail fetches an account using its email
func (s *Store) GetAccountByEmail(ctx context.Context, email string) (bookstore.Account, error) {
	var account bookstore.Account
	err := s.db.GetContext(ctx, &account, `SELECT * FROM account WHERE email = $1 LIMIT 1`, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = bookstore.NewNoResultError("account.email", err)
		}
		return bookstore.Account{}, fmt.Errorf("selecting account.email=%v: %w", email, err)
	}
	return account, nil
}

// ListAccounts returns a list of accounts
// to paginate, use the last received Account.ID to paginate
func (s *Store) ListAccounts(ctx context.Context, limit int, after uuid.UUID) ([]bookstore.Account, error) {
	accounts := make([]bookstore.Account, 0, limit)
	var err error
	if after != uuid.Nil {
		//if after uuid is provided, we add WHERE created_at > after via sub query to perform pagination
		//we use COALESCE to trigger a function that raises error if the selected ID does not exist
		query := `SELECT * FROM account WHERE created_at > COALESCE((SELECT created_at FROM account WHERE id = $2),raise_error_tz('Nonexistent UUID')) ORDER BY created_at LIMIT $1`
		err = s.db.SelectContext(ctx, &accounts, query, limit, after)
	} else {
		err = s.db.SelectContext(ctx, &accounts, `SELECT * FROM account ORDER BY created_at LIMIT $1`, limit)
	}
	err = enrichListPQError(err, "account")

	if err != nil {
		return nil, fmt.Errorf("listing accounts limit=%v after=%s: %w", limit, after, err)
	}
	return accounts, nil
}

// UpdateAccount updates the provided account using its ID
// note that CreatedAt, UpdatedAt cannot be set
func (s *Store) UpdateAccount(ctx context.Context, account bookstore.Account) error {
	if account.ID == uuid.Nil {
		return bookstore.ErrMissingID
	}
	var res sql.Result
	var err error
	if account.PasswordHash == "" {
		//if password hash is empty, we don't update it
		res, err = s.db.ExecContext(ctx, `UPDATE account SET name = $1,email = $2,is_admin = $4  WHERE id = $2`,
			account.Name, account.Email, account.Admin, account.ID)
	} else {
		res, err = s.db.ExecContext(ctx, `UPDATE account SET name = $1,email = $2,password_hash = $3,is_admin = $4  WHERE id = $2`,
			account.Name, account.Email, account.PasswordHash, account.Admin, account.ID)
	}

	if err != nil {
		err = enrichPQError(err, "account.name")
		return fmt.Errorf("updating account: %w", err)
	}
	err = checkAffectedRows(res, bookstore.NewNoResultError("account", err))
	if err != nil {
		return fmt.Errorf("updating account=%v: %w", account.ID, err)
	}
	return nil
}

// DeleteAccount deletes the specified account using its ID
func (s *Store) DeleteAccount(ctx context.Context, accountID uuid.UUID) error {
	if accountID == uuid.Nil {
		return fmt.Errorf("missing account id")
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM account WHERE id = $1`, accountID)
	if err != nil {
		err = enrichDeletePQError(err, "account")
		return fmt.Errorf("deleting account.id=%v: %w", accountID, err)
	}
	err = checkAffectedRows(res, bookstore.NewNoResultError("account", err))
	if err != nil {
		return fmt.Errorf("deleting account=%v: %w", accountID, err)
	}
	return nil
}
