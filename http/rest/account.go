package rest

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/thunder33345/bookstore"
	"github.com/thunder33345/bookstore/auth"
)

// CreateAccount handles signup
func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	data := &AccountRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	account := *data.Account
	var err error
	//we hash the password first, since incoming request contains the plain password
	account.PasswordHash, err = h.auth.Hash(account.PasswordHash)
	if err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
	}

	created, err := h.store.CreateAccount(r.Context(), account)

	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	render.Status(r, http.StatusOK)
	_ = render.Render(w, r, NewAccountResponse(created))
}

// GetAccount returns account info for current session
func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	ses, ok := auth.GetSession(r.Context())
	if !ok {
		//we try to get the session from auth,
		//but if we didn't get that, it means the account isn't logged in
		_ = render.Render(w, r, ErrUnauthorized)
		return
	}

	account, err := h.store.GetAccount(r.Context(), ses.ID)

	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	if err := render.Render(w, r, NewUserAccountResponse(account)); err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
}

// UpdateAccount updates current account's data
func (h *Handler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	ses, ok := auth.GetSession(r.Context())
	if !ok {
		_ = render.Render(w, r, ErrUnauthorized)
		return
	}

	data := &AccountRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	account := *data.Account
	account.ID = ses.ID
	//we disallow updating password directly
	account.PasswordHash = ""

	err := h.store.UpdateAccount(r.Context(), account)
	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateAccountPassword allows user to update their account's password
// this performs old password validation, unlike UpdateUser
func (h *Handler) UpdateAccountPassword(w http.ResponseWriter, r *http.Request) {
	ses, ok := auth.GetSession(r.Context())
	if !ok {
		_ = render.Render(w, r, ErrUnauthorized)
		return
	}
	data := &PasswordUpdateRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	dbAcc, err := h.store.GetAccount(r.Context(), ses.ID)
	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	//we validate the password hash matches the old password
	//before allowing password to be changed
	if ok, err := h.auth.Validate(dbAcc.PasswordHash, data.OldPassword); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	} else if !ok {
		_ = render.Render(w, r, ErrInvalidRequest(fmt.Errorf("invalid credentials")))
		return
	}

	dbAcc.PasswordHash, err = h.auth.Hash(data.NewPassword)
	if err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
	}

	err = h.store.UpdateAccount(r.Context(), dbAcc)
	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateAccountSession creates a session token using login credentials
// yes this is basically the endpoint for logging in
func (h *Handler) CreateAccountSession(w http.ResponseWriter, r *http.Request) {
	data := &SessionCreateRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	acc, err := h.store.GetAccountByEmail(r.Context(), data.Email)
	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}
	valid, err := h.auth.Validate(acc.PasswordHash, data.Password)
	if err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	if !valid {
		_ = render.Render(w, r, ErrInvalidRequest(fmt.Errorf("invalid credentials")))
		return
	}

	tok := h.auth.CreateSession(acc)

	render.Status(r, http.StatusOK)
	_ = render.Render(w, r, NewSessionCreateResponse(tok, acc))
}

// DeleteAccountSession removes current active session token(aka log out)
// using parameter all=true will log out all active session for current user
func (h *Handler) DeleteAccountSession(w http.ResponseWriter, r *http.Request) {
	ses, tok, ok := auth.GetSessionAndToken(r.Context())
	if !ok {
		_ = render.Render(w, r, ErrUnauthorized)
		return
	}

	all, err := strconv.ParseBool(r.URL.Query().Get("all"))
	if err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
	}

	if all {
		h.auth.DeleteSessionFor(ses.ID)
	} else {
		h.auth.DeleteSession(tok)
	}
	w.WriteHeader(http.StatusNoContent)
}

type UserAccountRequest struct {
	*bookstore.Account

	ProtectedID        uuid.UUID `json:"id"`
	ProtectedAdmin     bool      `json:"admin"`
	ProtectedCreatedAt time.Time `json:"created_at"`
	ProtectedUpdatedAt time.Time `json:"updated_at"`
}

func (a *UserAccountRequest) Bind(_ *http.Request) error {
	if a.Account == nil {
		return errors.New("missing required account fields")
	}

	a.ProtectedID = uuid.Nil
	a.ProtectedCreatedAt = time.Time{}
	a.ProtectedUpdatedAt = time.Time{}
	return nil
}

type UserAccountResponse struct {
	*bookstore.Account
	ProtectedHash string `json:"password"`
}

func NewUserAccountResponse(account bookstore.Account) *UserAccountResponse {
	resp := &UserAccountResponse{Account: &account}
	return resp
}

func (ur *UserAccountResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	ur.ProtectedHash = ""
	return nil
}

type SessionCreateRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *SessionCreateRequest) Bind(_ *http.Request) error {
	return nil
}

type SessionCreateResponse struct {
	Token   string            `json:"token"`
	Account bookstore.Account `json:"account"`
}

func NewSessionCreateResponse(token string, account bookstore.Account) *SessionCreateResponse {
	resp := &SessionCreateResponse{Token: token, Account: account}
	return resp
}

func (sc *SessionCreateResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	sc.Account.PasswordHash = ""
	return nil
}

type PasswordUpdateRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (a *PasswordUpdateRequest) Bind(_ *http.Request) error {
	return nil
}
