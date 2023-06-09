BEGIN;

CREATE TABLE session
(
    token      text NOT NULL PRIMARY KEY,
    account_id uuid NOT NULL,
    created_at timestamptz DEFAULT now(),
    CONSTRAINT fk_account FOREIGN KEY (account_id) REFERENCES account (id) ON DELETE CASCADE
);

COMMIT;