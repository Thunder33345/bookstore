package rest

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/thunder33345/bookstore"
)

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(ctxUUIDKey).(uuid.UUID)

	account, err := h.store.GetAccount(r.Context(), id)
	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	if err := render.Render(w, r, NewAccountResponse(account)); err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit := r.Context().Value(ctxKeyLimit).(int)
	after := r.Context().Value(ctxKeyAfter).(uuid.UUID)

	accounts, err := h.store.ListAccounts(r.Context(), limit, after)

	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	if err := render.RenderList(w, r, NewListAccountResponse(accounts)); err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(ctxUUIDKey).(uuid.UUID)

	data := &AccountRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	account := *data.Account
	account.ID = id
	//nil the password field to ignore updating it
	//use UpdateUserPassword to change that instead
	account.PasswordHash = ""

	err := h.store.UpdateAccount(r.Context(), account)
	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(ctxUUIDKey).(uuid.UUID)
	data := &AccountPasswordRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	//we fetch the data first
	account, err := h.store.GetAccount(r.Context(), id)
	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	//we hash the given password since incoming request is the plain password
	account.PasswordHash, err = h.auth.Hash(data.Password)
	if err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
	}

	err = h.store.UpdateAccount(r.Context(), account)
	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(ctxUUIDKey).(uuid.UUID)

	err := h.store.DeleteAccount(r.Context(), id)

	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}

func (h *Handler) DeleteUserSessions(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(ctxUUIDKey).(uuid.UUID)
	h.auth.DeleteSessionFor(id)
	w.WriteHeader(http.StatusNoContent)
}

type AccountRequest struct {
	*bookstore.Account

	ProtectedID        uuid.UUID `json:"id"`
	ProtectedCreatedAt time.Time `json:"created_at"`
	ProtectedUpdatedAt time.Time `json:"updated_at"`
}

func (a *AccountRequest) Bind(_ *http.Request) error {
	if a.Account == nil {
		return errors.New("missing required account fields")
	}

	a.ProtectedID = uuid.Nil
	a.ProtectedCreatedAt = time.Time{}
	a.ProtectedUpdatedAt = time.Time{}
	return nil
}

type AccountPasswordRequest struct {
	Password string `json:"password"`
}

func (a *AccountPasswordRequest) Bind(_ *http.Request) error {
	return nil
}

type AccountResponse struct {
	*bookstore.Account
}

func NewAccountResponse(account bookstore.Account) *AccountResponse {
	resp := &AccountResponse{Account: &account}
	return resp
}

func (rd *AccountResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

func NewListAccountResponse(accounts []bookstore.Account) []render.Renderer {
	list := make([]render.Renderer, 0, len(accounts))
	for _, article := range accounts {
		list = append(list, NewAccountResponse(article))
	}
	return list
}
