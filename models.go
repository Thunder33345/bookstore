package bookstore

import (
	"time"

	"github.com/google/uuid"
)

type Book struct {
	ISBN        uuid.UUID `json:"isbn"`
	Serial      int       `json:"serial"`
	Title       string    `json:"title"`
	AuthorID    uuid.UUID `json:"author_id"`
	GenreID     uuid.UUID `json:"genre_id"`
	PublishYear int       `json:"publish_year"`
	Fiction     bool      `json:"fiction"`
	CoverURL    string    `json:"cover_url"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Genre struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Author struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Account struct {
	ID           uuid.UUID `json:"ID"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Admin        bool      `json:"admin"`
	PasswordHash string    `json:"-"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Session struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Token  string

	CreatedAt time.Time `db:"created_at"`
}
