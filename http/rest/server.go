package rest

import (
	"context"
	"io"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/thunder33345/bookstore"
)

// Handler is the REST handler struct
// this stores the dependencies and some configuration
type Handler struct {
	store             store
	cover             coverStore
	defaultListLimit  int
	maxListLimit      int
	ignoreInvalidIBSN bool
}

// NewHandler creates a new Handler with given parameters
func NewHandler(store store, cover coverStore, options ...Option) *Handler {
	h := Handler{
		store:            store,
		cover:            cover,
		defaultListLimit: 50,
		maxListLimit:     100,
	}
	for _, option := range options {
		h = option(h)
	}
	return &h
}

// Mount will mount the whole rest handlers onto the given chi router
func (h *Handler) Mount(r chi.Router) {
	r.Route("/genres", func(r chi.Router) {
		r.With(h.PaginationLimitMiddleware, h.PaginationUUIDMiddleware).Get("/", h.ListGenres)
		r.Post("/", h.CreateGenre)
		r.With(UUIDCtx).Route("/{uuid}", func(r chi.Router) {
			r.Get("/", h.GetGenre)
			r.Put("/", h.UpdateGenre)
			r.Delete("/", h.DeleteGenre)
		})
	})

	r.Route("/authors", func(r chi.Router) {
		r.With(h.PaginationLimitMiddleware, h.PaginationUUIDMiddleware).Get("/", h.ListAuthors)
		r.Post("/", h.CreateAuthor)
		r.With(UUIDCtx).Route("/{uuid}", func(r chi.Router) {
			r.Get("/", h.GetAuthor)
			r.Put("/", h.UpdateAuthor)
			r.Delete("/", h.DeleteAuthor)
		})
	})

	r.Route("/books", func(r chi.Router) {
		r.With(h.PaginationLimitMiddleware, h.PaginationIBSNMiddleware).Get("/", h.ListBooks)
		r.With(ISBNCtx).Route("/{isbn}", func(r chi.Router) {
			r.Post("/", h.CreateBook)
			r.Get("/", h.GetBook)
			r.Put("/", h.UpdateBook)
			r.Delete("/", h.DeleteBook)
			r.Put("/cover", h.UpdateBookCover)
			r.Delete("/cover", h.DeleteBookCover)
		})
	})
}

// store is an interface of the DB
// using interfaces allow us to create a loose coupling with the database
type store interface {
	Init() error
	CreateGenre(ctx context.Context, genre bookstore.Genre) (bookstore.Genre, error)
	GetGenre(ctx context.Context, genreID uuid.UUID) (bookstore.Genre, error)
	ListGenres(ctx context.Context, limit int, after uuid.UUID) ([]bookstore.Genre, error)
	UpdateGenre(ctx context.Context, genre bookstore.Genre) error
	DeleteGenre(ctx context.Context, genreID uuid.UUID) error
	CreateAuthor(ctx context.Context, author bookstore.Author) (bookstore.Author, error)
	GetAuthor(ctx context.Context, authorID uuid.UUID) (bookstore.Author, error)
	ListAuthors(ctx context.Context, limit int, after uuid.UUID) ([]bookstore.Author, error)
	UpdateAuthor(ctx context.Context, author bookstore.Author) error
	DeleteAuthor(ctx context.Context, authorID uuid.UUID) error
	CreateBook(ctx context.Context, book bookstore.Book) (bookstore.Book, error)
	GetBook(ctx context.Context, bookID string) (bookstore.Book, error)
	ListBooks(ctx context.Context, limit int, after string, genresId []uuid.UUID, authorsId []uuid.UUID, searchTitle string) ([]bookstore.Book, error)
	UpdateBook(ctx context.Context, book bookstore.Book) error
	DeleteBook(ctx context.Context, bookID string) error
}

// coverStore is a minimal interface of fs.Store
type coverStore interface {
	StoreCover(ctx context.Context, bookID string, img io.ReadSeeker) error
	RemoveCover(ctx context.Context, bookID string) error
	ResolveCover(coverFile string) string
}
