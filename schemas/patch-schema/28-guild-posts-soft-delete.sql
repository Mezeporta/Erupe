BEGIN;

-- Add soft-delete column to guild_posts, matching the pattern used by characters and mail tables.
ALTER TABLE guild_posts ADD COLUMN IF NOT EXISTS deleted boolean DEFAULT false NOT NULL;

COMMIT;
