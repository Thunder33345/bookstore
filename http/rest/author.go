package rest

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/thunder33345/bookstore"
)

func (h *Handler) CreateAuthor(w http.ResponseWriter, r *http.Request) {
	data := &AuthorRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	author := *data.Author
	created, err := h.store.CreateAuthor(r.Context(), author)

	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	render.Status(r, http.StatusOK)
	_ = render.Render(w, r, NewAuthorResponse(created))
}

func (h *Handler) GetAuthor(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(ctxUUIDKey).(uuid.UUID)
	if id == uuid.Nil {
		return
	}

	author, err := h.store.GetAuthor(r.Context(), id)

	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	if err := render.Render(w, r, NewAuthorResponse(author)); err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
}

func (h *Handler) ListAuthors(w http.ResponseWriter, r *http.Request) {
	limit := r.Context().Value(ctxKeyLimit).(int)
	after := r.Context().Value(ctxKeyAfter).(uuid.UUID)

	authors, err := h.store.ListAuthors(r.Context(), limit, after)

	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	if err := render.RenderList(w, r, NewListAuthorResponse(authors)); err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
}

func (h *Handler) UpdateAuthor(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(ctxUUIDKey).(uuid.UUID)

	data := &AuthorRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	author := *data.Author
	author.ID = id

	err := h.store.UpdateAuthor(r.Context(), author)
	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteAuthor(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(ctxUUIDKey).(uuid.UUID)

	err := h.store.DeleteAuthor(r.Context(), id)

	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}

type AuthorRequest struct {
	*bookstore.Author

	ProtectedID        uuid.UUID `json:"id"`
	ProtectedCreatedAt time.Time `json:"created_at"`
	ProtectedUpdatedAt time.Time `json:"updated_at"`
}

func (a *AuthorRequest) Bind(r *http.Request) error {
	if a.Author == nil {
		return errors.New("missing required author fields")
	}

	a.ProtectedID = uuid.Nil
	a.ProtectedCreatedAt = time.Time{}
	a.ProtectedUpdatedAt = time.Time{}
	return nil
}

type AuthorResponse struct {
	*bookstore.Author
}

func NewAuthorResponse(author bookstore.Author) *AuthorResponse {
	resp := &AuthorResponse{Author: &author}
	return resp
}

func (rd *AuthorResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewListAuthorResponse(authors []bookstore.Author) []render.Renderer {
	list := make([]render.Renderer, 0, len(authors))
	for _, article := range authors {
		list = append(list, NewAuthorResponse(article))
	}
	return list
}
