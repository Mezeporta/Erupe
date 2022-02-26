BEGIN;
CREATE TYPE guild_application_type AS ENUM ('applied', 'invited');

CREATE TABLE guild_applications
(
    id               serial                 NOT NULL PRIMARY KEY,
    guild_id         int                    NOT NULL REFERENCES guilds (id),
    character_id     int                    NOT NULL REFERENCES characters (id),
    actor_id         int                    NOT NULL REFERENCES characters (id),
    application_type guild_application_type NOT NULL,
    created_at       timestamp              NOT NULL DEFAULT now(),
    CONSTRAINT guild_application_character_id UNIQUE (guild_id, character_id)
);

CREATE INDEX guild_application_type_index ON guild_applications (application_type);

ALTER TABLE guild_characters
    DROP COLUMN is_applicant;

ALTER TABLE guild_characters
    RENAME COLUMN is_sub_leader TO avoid_leadership;

ALTER TABLE guilds
    ALTER COLUMN main_motto SET DEFAULT 0;

ALTER TABLE guilds
    ADD COLUMN icon      bytea,
    ADD COLUMN sub_motto int DEFAULT 0,
    ALTER COLUMN main_motto TYPE int USING 0;
END;