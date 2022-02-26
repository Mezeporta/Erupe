BEGIN;

CREATE TABLE guilds
(
    id         serial NOT NULL PRIMARY KEY,
    name       varchar(24),
    created_at timestamp    DEFAULT NOW(),
    leader_id  int    NOT NULL,
    main_motto    varchar(255) DEFAULT ''
);

CREATE TABLE guild_characters
(
    id           serial NOT NULL PRIMARY KEY,
    guild_id     bigint REFERENCES guilds (id),
    character_id bigint REFERENCES characters (id),
    joined_at    timestamp DEFAULT NOW()
);

CREATE UNIQUE INDEX guild_character_unique_index ON guild_characters (character_id);

END;