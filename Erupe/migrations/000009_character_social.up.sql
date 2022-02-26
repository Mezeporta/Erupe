BEGIN;

ALTER TABLE characters
    ADD COLUMN restrict_guild_scout bool NOT NULL DEFAULT false;

END;