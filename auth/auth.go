package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/puzpuzpuz/xsync"
	"github.com/thunder33345/bookstore"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	ses    xsync.MapOf[string, bookstore.Session]
	db     db
	pwCost int
}

func NewAuth(db db, pwCost int) *Auth {
	return &Auth{
		ses:    xsync.MapOf[string, bookstore.Session]{},
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

func (a *Auth) GetSession(token string) (bookstore.Session, bool) {
	return a.ses.Load(token)
}

func (a *Auth) CreateSession(account bookstore.Account) string {
	sessionToken := uuid.NewString()
	a.ses.Store(sessionToken, bookstore.Session{Account: account})
	return sessionToken
}

func (a *Auth) DeleteSession(token string) {
	a.ses.Delete(token)
}

func (a *Auth) DeleteSessionFor(user uuid.UUID) {
	a.ses.Range(func(key string, s bookstore.Session) bool {
		if s.ID == user {
			a.ses.Delete(key)
		}
		return true
	})
}

type db interface {
	GetAccountByEmail(ctx context.Context, email string) (bookstore.Account, error)
}
