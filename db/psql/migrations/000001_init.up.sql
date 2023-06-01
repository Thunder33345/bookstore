BEGIN;

CREATE TABLE authors
(
    id           uuid NOT NULL PRIMARY KEY,
    author       text NOT NULL UNIQUE CHECK (author <> ''),
    created_at   timestamp DEFAULT current_timestamp,
    last_updated timestamp DEFAULT current_timestamp
);

CREATE TABLE genres
(
    id           uuid NOT NULL PRIMARY KEY,
    genre        text NOT NULL UNIQUE CHECK (genre <> ''),
    created_at   timestamp DEFAULT current_time,
    last_updated timestamp DEFAULT current_date
);
CREATE TABLE books
(
    isbn         text PRIMARY KEY,
    title        text    NOT NULL CHECK (title <> ''),
    publish_year integer NOT NULL CHECK (publish_year > 0),
    has_cover    boolean   DEFAULT false,
    author_id    uuid    NOT NULL,
    genre_id     uuid    NOT NULL,
    last_updated timestamp DEFAULT current_timestamp,
    created_at   timestamp DEFAULT current_timestamp,
    CONSTRAINT fk_author FOREIGN KEY (author_id) REFERENCES authors (id) ON DELETE RESTRICT,
    CONSTRAINT fk_genre FOREIGN KEY (genre_id) REFERENCES genres (id) ON DELETE RESTRICT
);

CREATE TABLE users
(
    id            uuid        NOT NULL,
    name          text        NOT NULL CHECK (name <> ''),
    email         text UNIQUE NOT NULL CHECK (email <> ''),
    password_hash text        NOT NULL CHECK (password_hash <> ''),
    is_admin      boolean   DEFAULT false,
    created_at    timestamp DEFAULT current_date,
    last_updated  timestamp DEFAULT current_date,
    CONSTRAINT pk_users_id PRIMARY KEY (id)
);

CREATE TABLE sessions
(
    user_id    uuid NOT NULL,
    token      text PRIMARY KEY UNIQUE NOT NULL CHECK (token <> ''),
    created_at timestamp DEFAULT current_date,
    CONSTRAINT fk_users FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

COMMIT;