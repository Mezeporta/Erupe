BEGIN;

ALTER TABLE characters
    DROP COLUMN restrict_guild_scout;

END;