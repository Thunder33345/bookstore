BEGIN;

CREATE TABLE cover_data
(
    isbn       text PRIMARY KEY NOT NULL,
    cover_file text CHECK (cover_file <> ''),

    updated_at timestamptz DEFAULT now(),
    created_at timestamptz DEFAULT now(),
    CONSTRAINT fk_isbn FOREIGN KEY (isbn) REFERENCES book (isbn) ON DELETE CASCADE
);

INSERT INTO cover_data(isbn, cover_file)
SELECT isbn, cover_file
FROM book
WHERE cover_file != '';

CREATE TRIGGER trigger_update_timestamp
    BEFORE UPDATE
    ON cover_data
    FOR EACH ROW
EXECUTE PROCEDURE sync_updated_at();

ALTER TABLE book
    DROP COLUMN cover_file;

COMMIT;