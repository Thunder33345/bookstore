package bookstore

import (
	"time"

	"github.com/google/uuid"
)

type Book struct {
	ISBN        string    `json:"isbn"`
	Title       string    `json:"title"`
	AuthorID    uuid.UUID `json:"author_id" db:"author_id"`
	GenreID     uuid.UUID `json:"genre_id" db:"genre_id"`
	PublishYear int       `json:"publish_year" db:"publish_year"`
	Fiction     bool      `json:"fiction"`
	CoverURL    string    `json:"cover_url" db:"cover_file"`

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
	PasswordHash string    `json:"password,omitempty"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Session embeds Account
// mostly for future proofing and distinction
type Session struct {
	Account
}
