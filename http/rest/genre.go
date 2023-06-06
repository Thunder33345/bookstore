package rest

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/thunder33345/bookstore"
)

func (h *Handler) CreateGenre(w http.ResponseWriter, r *http.Request) {
	data := &GenreRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	genre := *data.Genre
	created, err := h.store.CreateGenre(r.Context(), genre)

	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	render.Status(r, http.StatusOK)
	_ = render.Render(w, r, NewGenreResponse(created))
}

func (h *Handler) GetGenre(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(ctxUUIDKey).(uuid.UUID)

	genre, err := h.store.GetGenre(r.Context(), id)

	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	if err := render.Render(w, r, NewGenreResponse(genre)); err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
}

func (h *Handler) ListGenres(w http.ResponseWriter, r *http.Request) {
	limit := r.Context().Value(ctxKeyLimit).(int)
	after := r.Context().Value(ctxKeyAfter).(uuid.UUID)

	genres, err := h.store.ListGenres(r.Context(), limit, after)

	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	if err := render.RenderList(w, r, NewListGenreResponse(genres)); err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
}

func (h *Handler) UpdateGenre(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(ctxUUIDKey).(uuid.UUID)

	data := &GenreRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	genre := *data.Genre
	genre.ID = id

	err := h.store.UpdateGenre(r.Context(), genre)
	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteGenre(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(ctxUUIDKey).(uuid.UUID)

	err := h.store.DeleteGenre(r.Context(), id)

	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}

type GenreRequest struct {
	*bookstore.Genre

	ProtectedID        uuid.UUID `json:"id"`
	ProtectedCreatedAt time.Time `json:"created_at"`
	ProtectedUpdatedAt time.Time `json:"updated_at"`
}

func (a *GenreRequest) Bind(r *http.Request) error {
	if a.Genre == nil {
		return errors.New("missing required genre fields")
	}

	a.ProtectedID = uuid.Nil
	a.ProtectedCreatedAt = time.Time{}
	a.ProtectedUpdatedAt = time.Time{}
	return nil
}

type GenreResponse struct {
	*bookstore.Genre
}

func NewGenreResponse(genre bookstore.Genre) *GenreResponse {
	resp := &GenreResponse{Genre: &genre}
	return resp
}

func (rd *GenreResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewListGenreResponse(genres []bookstore.Genre) []render.Renderer {
	list := make([]render.Renderer, 0, len(genres))
	for _, article := range genres {
		list = append(list, NewGenreResponse(article))
	}
	return list
}
