BEGIN;

CREATE TYPE festival_colour AS ENUM ('none', 'red', 'blue');

ALTER TABLE guilds
    ADD COLUMN comment         varchar(255) NOT NULL DEFAULT '',
    ADD COLUMN festival_colour festival_colour DEFAULT 'none',
    ADD COLUMN guild_hall      int DEFAULT 0;


END;