-- Rotating savedata backup table (3 slots per character, time-gated).
-- Prevents permanent data loss from save corruption by keeping recent snapshots.
CREATE TABLE IF NOT EXISTS savedata_backups (
    char_id     INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    slot        SMALLINT NOT NULL CHECK (slot BETWEEN 0 AND 2),
    savedata    BYTEA NOT NULL,
    saved_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (char_id, slot)
);
