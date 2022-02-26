BEGIN;

ALTER TABLE characters
    ADD COLUMN exp uint16,
    ADD COLUMN weapon uint16,
    ADD COLUMN last_login integer;

END;