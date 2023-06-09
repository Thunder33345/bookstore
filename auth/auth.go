package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/thanhpk/randstr"
	"github.com/thunder33345/bookstore"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	ses    session
	db     db
	pwCost int
}

func NewAuth(session session, db db, pwCost int) *Auth {
	return &Auth{
		ses:    session,
		db:     db,
		pwCost: pwCost,
	}
}

func (a *Auth) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), a.pwCost)
	return string(hash), err
}

func (a *Auth) Validate(hash, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (a *Auth) GetSession(ctx context.Context, token string) (bookstore.Session, error) {
	return a.ses.GetSession(ctx, token)
}

func (a *Auth) CreateSession(ctx context.Context, account bookstore.Account) (string, error) {
	sessionToken := randstr.Base62(16)
	err := a.ses.StoreSession(ctx, sessionToken, account)
	if err != nil {
		return "", err
	}
	return sessionToken, err
}

func (a *Auth) DeleteSession(ctx context.Context, token string) error {
	return a.ses.DeleteSession(ctx, token)
}

func (a *Auth) DeleteSessionFor(ctx context.Context, user uuid.UUID) error {
	return a.ses.DeleteSessionsFor(ctx, user)
}

type db interface {
	GetAccountByEmail(ctx context.Context, email string) (bookstore.Account, error)
}
type session interface {
	GetSession(ctx context.Context, token string) (bookstore.Session, error)
	StoreSession(ctx context.Context, token string, account bookstore.Account) error
	DeleteSession(ctx context.Context, token string) error
	DeleteSessionsFor(ctx context.Context, accountID uuid.UUID) error
}
