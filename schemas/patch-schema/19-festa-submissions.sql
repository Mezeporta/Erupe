CREATE TABLE IF NOT EXISTS festa_submissions (
    character_id int NOT NULL,
    guild_id int NOT NULL,
    trial_type int NOT NULL,
    souls int NOT NULL,
    timestamp timestamp with time zone NOT NULL
);

ALTER TABLE guild_characters DROP COLUMN IF EXISTS souls;

DO $$ BEGIN
    ALTER TYPE festival_colour RENAME TO festival_color;
EXCEPTION
    WHEN undefined_object THEN NULL;
    WHEN duplicate_object THEN NULL;
END $$;