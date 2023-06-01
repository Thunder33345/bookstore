BEGIN;

CREATE TABLE author
(
    id          uuid        NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    name text        NOT NULL UNIQUE CHECK (name <> ''),
    created_at  timestamptz NOT NULL             DEFAULT now(),
    updated_at  timestamptz NOT NULL             DEFAULT now()
);
-- Create index for author.created_at, as we will be using that for paging
CREATE UNIQUE INDEX index_author ON author USING btree (created_at ASC);

CREATE TABLE genre
(
    id         uuid NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       text NOT NULL UNIQUE CHECK (name <> ''),
    created_at timestamptz               DEFAULT now(),
    updated_at timestamptz               DEFAULT now()
);
CREATE UNIQUE INDEX index_genre ON genre USING btree (created_at ASC);

CREATE TABLE book
(
    isbn         text PRIMARY KEY NOT NULL CHECK (isbn <> ''),
    title        text             NOT NULL CHECK (title <> ''),
    publish_year integer          NOT NULL CHECK (publish_year > 0),
    has_cover    boolean     DEFAULT false,
    author_id    uuid             NOT NULL,
    genre_id     uuid             NOT NULL,
    updated_at   timestamptz DEFAULT now(),
    created_at   timestamptz DEFAULT now(),
    CONSTRAINT fk_author FOREIGN KEY (author_id) REFERENCES author (id) ON DELETE RESTRICT,
    CONSTRAINT fk_genre FOREIGN KEY (genre_id) REFERENCES genre (id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX index_book ON book USING btree (created_at ASC);

-- Using user seems to be not recommended
CREATE TABLE account
(
    id            uuid        NOT NULL,
    name          text        NOT NULL CHECK (name <> ''),
    email         text UNIQUE NOT NULL CHECK (email <> ''),
    password_hash text        NOT NULL CHECK (password_hash <> ''),
    is_admin      boolean     DEFAULT false,
    created_at    timestamptz DEFAULT now(),
    updated_at    timestamptz DEFAULT now(),
    CONSTRAINT pk_users_id PRIMARY KEY (id)
);
CREATE UNIQUE INDEX index_account ON account USING btree (created_at ASC);

CREATE TABLE sessions
(
    account_id    uuid                    NOT NULL,
    token      text PRIMARY KEY UNIQUE NOT NULL CHECK (token <> ''),
    created_at timestamptz DEFAULT now(),
    CONSTRAINT fk_account FOREIGN KEY (account_id) REFERENCES account (id) ON DELETE CASCADE
);

COMMIT;