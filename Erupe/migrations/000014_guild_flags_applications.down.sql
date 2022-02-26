BEGIN;
ALTER TABLE guild_characters
    RENAME COLUMN avoid_leadership TO is_sub_leader;

ALTER TABLE guild_characters
    ADD COLUMN is_applicant bool NOT NULL DEFAULT false;

ALTER TABLE guilds
    DROP COLUMN icon,
    ALTER COLUMN main_motto TYPE varchar USING '',
    DROP COLUMN sub_motto;

ALTER TABLE guilds
    ALTER COLUMN main_motto SET DEFAULT '';

DROP TABLE guild_applications;
DROP TYPE guild_application_type;
END;