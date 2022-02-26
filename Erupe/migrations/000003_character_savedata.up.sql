BEGIN;

ALTER TABLE characters
    ADD COLUMN savedata bytea;

END;