-- Save transfer tokens: one-time admin-granted permission for a character
-- to receive an imported save via the API endpoint.
-- NULL means no import is pending for this character.
ALTER TABLE characters
    ADD COLUMN IF NOT EXISTS savedata_import_token        TEXT,
    ADD COLUMN IF NOT EXISTS savedata_import_token_expiry TIMESTAMPTZ;
