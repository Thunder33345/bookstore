package rest

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/thunder33345/bookstore"
)

func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(ctxISBNKey).(string)

	data := &BookRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrInvalidRequestBody(err))
		return
	}

	book := *data.Book

	var err error
	book.ISBN, err = h.validateISBN(id)
	if err != nil {
		_ = render.Render(w, r, ErrInvalidIDRequest(err))
		return
	}

	created, err := h.store.CreateBook(r.Context(), book)
	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	render.Status(r, http.StatusOK)
	_ = render.Render(w, r, NewBookResponse(created, h.cover.ResolveCover))
}

func (h *Handler) GetBook(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(ctxISBNKey).(string)

	book, err := h.store.GetBook(r.Context(), id)

	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	if err := render.Render(w, r, NewBookResponse(book, h.cover.ResolveCover)); err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
}

func (h *Handler) ListBooks(w http.ResponseWriter, r *http.Request) {
	limit := r.Context().Value(ctxKeyLimit).(int)
	after := r.Context().Value(ctxKeyAfter).(string)

	err := r.ParseForm()
	if err != nil {
		_ = render.Render(w, r, ErrInvalidRequestBody(err))
		return
	}

	genreIds, err := stringSliceToUUID(r.Form["genre"])
	if err != nil {
		_ = render.Render(w, r, ErrInvalidRequestParam("genre", err))
		return
	}

	authorIds, err := stringSliceToUUID(r.Form["author"])
	if err != nil {
		_ = render.Render(w, r, ErrInvalidRequestParam("author", err))
		return
	}

	books, err := h.store.ListBooks(r.Context(), limit, after, genreIds, authorIds, r.URL.Query().Get("name"))

	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	if err := render.RenderList(w, r, NewListBookResponse(books, h.cover.ResolveCover)); err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
}

func (h *Handler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(ctxISBNKey).(string)

	data := &BookRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrInvalidRequestBody(err))
		return
	}
	book := *data.Book
	book.ISBN = id

	err := h.store.UpdateBook(r.Context(), book)
	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(ctxISBNKey).(string)

	err := h.store.DeleteBook(r.Context(), id)

	if err != nil {
		_ = render.Render(w, r, ErrQueryResponse(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}

type BookRequest struct {
	*bookstore.Book

	ProtectedISBN      string    `json:"isbn"`
	ProtectedCoverURL  string    `json:"cover_url"`
	ProtectedCreatedAt time.Time `json:"created_at"`
	ProtectedUpdatedAt time.Time `json:"updated_at"`
}

func (b *BookRequest) Bind(r *http.Request) error {
	if b.Book == nil {
		return errors.New("missing required book fields")
	}
	b.ProtectedISBN = ""
	b.ProtectedCoverURL = ""
	b.ProtectedCreatedAt = time.Time{}
	b.ProtectedUpdatedAt = time.Time{}

	return nil
}

type coverResolver func(coverFile string) string
type BookResponse struct {
	*bookstore.Book
	resolver coverResolver
}

func NewBookResponse(book bookstore.Book, coverResolver coverResolver) *BookResponse {
	resp := &BookResponse{Book: &book, resolver: coverResolver}
	return resp
}

func (b *BookResponse) Render(w http.ResponseWriter, r *http.Request) error {
	b.Book.CoverURL = b.resolver(b.Book.CoverURL)
	return nil
}

func NewListBookResponse(books []bookstore.Book, resolver coverResolver) []render.Renderer {
	list := make([]render.Renderer, 0, len(books))
	for _, article := range books {
		list = append(list, NewBookResponse(article, resolver))
	}
	return list
}
