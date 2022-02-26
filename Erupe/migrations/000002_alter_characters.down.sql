BEGIN;

ALTER TABLE characters
    DROP COLUMN exp,
    DROP COLUMN weapon,
    DROP COLUMN last_login;

END;