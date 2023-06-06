package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/puzpuzpuz/xsync"
	"github.com/thunder33345/bookstore"
	"github.com/thunder33345/bookstore/http/rest"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	ses    xsync.MapOf[string, bookstore.Session]
	db     db
	pwCost int
}

var ErrInvalidCredentials = errors.New("invalid credentials")

func (a *Auth) Login(ctx context.Context, email string, password string) (string, error) {
	acc, err := a.db.GetAccountByEmail(ctx, email)
	if err != nil {
		var nr *bookstore.NoResultError
		if errors.As(err, &nr) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}
	valid, err := a.Validate(acc.PasswordHash, password)
	if err != nil {
		return "", err
	}
	if !valid {
		return "", ErrInvalidCredentials
	}

	return a.CreateSession(acc), nil
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

// AuthenticatedOnly is a middleware to enforce authentication
func (a *Auth) AuthenticatedOnly(requireAdmin bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//try to populate the session into context
			//so other handlers can also use them
			r, account, err := a.populateSession(r)
			if err != nil {
				_ = render.Render(w, r, rest.ErrUnauthorized)
				return
			}
			if requireAdmin && account.Admin == false {
				_ = render.Render(w, r, rest.ErrUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type ctxKey string

// populateSession tries to populate session data into context using header
func (a *Auth) populateSession(r *http.Request) (*http.Request, bookstore.Session, error) {
	//if it's already populated, we skip it
	//this could happen if middlewares got chained
	if ses, ok := GetSession(r.Context()); ok {
		return r, ses, nil
	}

	ah := r.Header.Get("Authorization")
	if !strings.HasPrefix(ah, "Bearer ") {
		return nil, bookstore.Session{}, ErrInvalidCredentials
	}
	ah = strings.TrimLeft(ah, "Bearer ")

	account, ok := a.ses.Load(ah)
	if !ok {
		return nil, bookstore.Session{}, ErrInvalidCredentials
	}
	r = r.WithContext(context.WithValue(r.Context(), ctxKey("user"), account))
	r = r.WithContext(context.WithValue(r.Context(), ctxKey("token"), ah))
	return r, account, nil
}

func GetSession(ctx context.Context) (bookstore.Session, bool) {
	ses, ok := ctx.Value(ctxKey("user")).(bookstore.Session)
	return ses, ok
}

func GetSessionToken(ctx context.Context) (string, bool) {
	tok, ok := ctx.Value(ctxKey("token")).(string)
	return tok, ok
}

func GetSessionAndToken(ctx context.Context) (bookstore.Session, string, bool) {
	ses, okU := ctx.Value(ctxKey("user")).(bookstore.Session)
	tok, okT := ctx.Value(ctxKey("token")).(string)
	return ses, tok, okU && okT
}

type db interface {
	GetAccountByEmail(ctx context.Context, email string) (bookstore.Account, error)
}
