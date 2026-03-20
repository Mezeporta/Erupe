-- Savedata integrity protections: rotating backups + checksum verification.

-- Rotating backup table (3 slots per character, time-gated).
-- Prevents permanent data loss from save corruption by keeping recent snapshots.
CREATE TABLE IF NOT EXISTS savedata_backups (
    char_id     INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    slot        SMALLINT NOT NULL CHECK (slot BETWEEN 0 AND 2),
    savedata    BYTEA NOT NULL,
    saved_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (char_id, slot)
);

-- SHA-256 checksum column for savedata integrity verification.
-- Stored as 32 raw bytes (not hex). NULL means no hash computed yet
-- (backwards-compatible with existing data).
ALTER TABLE characters ADD COLUMN IF NOT EXISTS savedata_hash BYTEA;
