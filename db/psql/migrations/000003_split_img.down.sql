BEGIN;

ALTER TABLE book
    ADD COLUMN cover_file text;

UPDATE book b
SET cover_file = d.cover_file
FROM cover_data d
WHERE b.isbn = d.isbn;

DROP TABLE cover_data;

COMMIT;