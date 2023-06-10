BEGIN;

CREATE TABLE cover_data
(
    isbn       text PRIMARY KEY NOT NULL,
    cover_file text CHECK (cover_file <> ''),
    file_type  text CHECK (cover_file <> ''),

    updated_at timestamptz DEFAULT now(),
    created_at timestamptz DEFAULT now(),
    CONSTRAINT fk_isbn FOREIGN KEY (isbn) REFERENCES book (isbn) ON DELETE CASCADE
);

INSERT INTO cover(isbn, cover_file)
SELECT isbn, cover_file
FROM book
WHERE cover_file != '';

CREATE TRIGGER trigger_update_timestamp
    BEFORE UPDATE
    ON cover
    FOR EACH ROW
EXECUTE PROCEDURE sync_updated_at();

ALTER TABLE book
    DROP COLUMN cover_file;

COMMIT;