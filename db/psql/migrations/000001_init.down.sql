BEGIN;
DROP TABLE IF EXISTS book;
DROP TABLE IF EXISTS genre;
DROP TABLE IF EXISTS author;

DROP TABLE IF EXISTS account;

DROP FUNCTION IF EXISTS sync_updated_at;
COMMIT;
