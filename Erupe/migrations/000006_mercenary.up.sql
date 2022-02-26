BEGIN;

ALTER TABLE characters
    ADD COLUMN savemercenary bytea;

END;