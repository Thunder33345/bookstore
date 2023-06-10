BEGIN;
DROP TABLE cover_data;

ALTER TABLE book
    ADD COLUMN cover_file text;

COMMIT;