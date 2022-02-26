BEGIN;

ALTER TABLE guilds
    DROP COLUMN comment,
    DROP COLUMN festival_colour,
    DROP COLUMN guild_hall;

DROP TYPE festival_colour;

END;