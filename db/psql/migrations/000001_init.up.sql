BEGIN;

CREATE TABLE author
(
    id         uuid        NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       text        NOT NULL UNIQUE CHECK (name <> ''),
    created_at timestamptz NOT NULL             DEFAULT now(),
    updated_at timestamptz NOT NULL             DEFAULT now()
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
    fiction      boolean          NOT NULL,
    cover_file   text,
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

-- create a function to raise errors, this is used if our sub-query fails
create or replace function raise_error_tz(text) returns timestamptz as
$$
begin
    raise exception '%', $1;
    return '1970-01-01'::timestamptz;
end;
$$ language plpgsql;

-- Create a trigger function to keep updated_at in sync
CREATE FUNCTION sync_updated_at() RETURNS trigger AS
$$
BEGIN
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply the trigger to all tables
CREATE TRIGGER trigger_update_timestamp
    BEFORE UPDATE
    ON author
    FOR EACH ROW
EXECUTE PROCEDURE sync_updated_at();

CREATE TRIGGER trigger_update_timestamp
    BEFORE UPDATE
    ON genre
    FOR EACH ROW
EXECUTE PROCEDURE sync_updated_at();

CREATE TRIGGER trigger_update_timestamp
    BEFORE UPDATE
    ON book
    FOR EACH ROW
EXECUTE PROCEDURE sync_updated_at();

CREATE TRIGGER trigger_update_timestamp
    BEFORE UPDATE
    ON account
    FOR EACH ROW
EXECUTE PROCEDURE sync_updated_at();

COMMIT;