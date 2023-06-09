package session

import (
	"context"

	"github.com/google/uuid"
	"github.com/puzpuzpuz/xsync"
	"github.com/thunder33345/bookstore"
)

type Memory struct {
	ses *xsync.MapOf[string, bookstore.Session]
}

func NewMemory() *Memory {
	sm := xsync.NewMapOf[bookstore.Session]()
	return &Memory{
		ses: sm,
	}
}

func (a *Memory) StoreSession(_ context.Context, token string, account bookstore.Account) error {
	a.ses.Store(token, bookstore.Session{Account: account})
	return nil
}

func (a *Memory) GetSession(_ context.Context, token string) (bookstore.Session, error) {
	ses, found := a.ses.Load(token)
	if !found {
		return bookstore.Session{}, bookstore.ErrInvalidSession
	}
	return ses, nil
}

func (a *Memory) DeleteSession(_ context.Context, token string) error {
	a.ses.Delete(token)
	return nil
}

func (a *Memory) DeleteSessionsFor(_ context.Context, accountID uuid.UUID) error {
	a.ses.Range(func(key string, s bookstore.Session) bool {
		if s.ID == accountID {
			a.ses.Delete(key)
		}
		//return true to keep iterating through the whole session
		return true
	})
	return nil
}
