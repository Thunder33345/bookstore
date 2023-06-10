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
	passwordvalidator "github.com/wagslane/go-password-validator"
)

// CreateAccount handles signup
func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	data := &AccountRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	account := *data.Account
	account.Admin = false

	if data.PasswordHash == "" {
		_ = render.Render(w, r, ErrInvalidRequest(fmt.Errorf("no password provided")))
		return
	}

	if err := passwordvalidator.Validate(data.PasswordHash, h.minPWEntropy); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

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
	tok, err := h.auth.CreateSession(r.Context(), created)
	if err != nil {
		_ = render.Render(w, r, ErrSessionResponse(err))
	}

	render.Status(r, http.StatusOK)
	_ = render.Render(w, r, NewSessionCreateResponse(tok, created))
}

// GetAccount returns account info for current session
func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	ses, ok := GetSession(r.Context())
	if !ok {
		//we try to get the session from context,
		//it should have been populated otherwise middleware wouldn't let the request through
		_ = render.Render(w, r, ErrSessionResponse(bookstore.ErrMissingSessionData))
		return
	}

	if err := render.Render(w, r, NewUserAccountResponse(ses.Account)); err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
}

// UpdateAccount updates current account's data
func (h *Handler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	ses, ok := GetSession(r.Context())
	if !ok {
		_ = render.Render(w, r, ErrSessionResponse(bookstore.ErrMissingSessionData))
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

	err := h.store.SafeUpdateAccount(r.Context(), account)
	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateAccountPassword allows user to update their account's password
// this performs old password validation, unlike UpdateUser
func (h *Handler) UpdateAccountPassword(w http.ResponseWriter, r *http.Request) {
	ses, ok := GetSession(r.Context())
	if !ok {
		_ = render.Render(w, r, ErrSessionResponse(bookstore.ErrMissingSessionData))
		return
	}
	data := &PasswordUpdateRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if data.OldPassword == data.NewPassword {
		_ = render.Render(w, r, ErrInvalidRequest(fmt.Errorf("same password provided")))
		return
	}
	if data.NewPassword == "" {
		_ = render.Render(w, r, ErrInvalidRequest(fmt.Errorf("no new password provided")))
		return
	}

	if err := passwordvalidator.Validate(data.NewPassword, h.minPWEntropy); err != nil {
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

	tok, err := h.auth.CreateSession(r.Context(), acc)
	if err != nil {
		_ = render.Render(w, r, ErrSessionResponse(err))
	}

	render.Status(r, http.StatusOK)
	_ = render.Render(w, r, NewSessionCreateResponse(tok, acc))
}

// DeleteAccountSession removes current active session token(aka log out)
// using parameter all=true will log out all active session for current user
func (h *Handler) DeleteAccountSession(w http.ResponseWriter, r *http.Request) {
	ses, tok, ok := GetSessionAndToken(r.Context())
	if !ok {
		_ = render.Render(w, r, ErrSessionResponse(bookstore.ErrMissingSessionData))
		return
	}
	var err error
	all := false

	if query := r.URL.Query().Get("all"); query != "" {
		all, err = strconv.ParseBool(query)
		if err != nil {
			_ = render.Render(w, r, ErrInvalidRequest(err))
			return
		}
	}

	if all {
		err = h.auth.DeleteSessionFor(r.Context(), ses.ID)
	} else {
		err = h.auth.DeleteSession(r.Context(), tok)
	}
	if err != nil {
		_ = render.Render(w, r, ErrSessionResponse(err))
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
	ProtectedHash string `json:"password,omitempty"`
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
