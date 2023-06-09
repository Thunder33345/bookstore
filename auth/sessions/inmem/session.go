package inmem

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/puzpuzpuz/xsync"
	"github.com/thunder33345/bookstore"
)

type Session struct {
	ses *xsync.MapOf[string, bookstore.Session]
}

func NewSession() *Session {
	sm := xsync.NewMapOf[bookstore.Session]()
	return &Session{
		ses: sm,
	}
}

func (a *Session) GetSession(_ context.Context, token string) (bookstore.Session, error) {
	ses, found := a.ses.Load(token)
	if !found {
		return bookstore.Session{}, errors.New("session not found")
	}
	return ses, nil
}

func (a *Session) StoreSession(_ context.Context, token string, account bookstore.Account) error {
	a.ses.Store(token, bookstore.Session{Account: account})
	return nil
}

func (a *Session) DeleteSession(_ context.Context, token string) error {
	a.ses.Delete(token)
	return nil
}

func (a *Session) DeleteSessionFor(_ context.Context, accountID uuid.UUID) error {
	a.ses.Range(func(key string, s bookstore.Session) bool {
		if s.ID == accountID {
			a.ses.Delete(key)
		}
		//return true to keep iterating through the whole session
		return true
	})
	return nil
}
