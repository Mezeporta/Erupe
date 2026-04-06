-- Diva Defense (United Defense) extended schema.
-- Adds bead selection, per-bead point accumulation, interception points,
-- and prize reward tables for personal and guild tracks.

-- Interception map data per guild (binary blob, existing column pattern).
ALTER TABLE guilds ADD COLUMN IF NOT EXISTS interception_maps bytea;

-- Per-character interception points keyed by quest file ID.
ALTER TABLE guild_characters ADD COLUMN IF NOT EXISTS interception_points jsonb NOT NULL DEFAULT '{}';

-- Prize reward table for personal and guild tracks.
CREATE TABLE IF NOT EXISTS diva_prizes (
    id         SERIAL PRIMARY KEY,
    type       VARCHAR(10) NOT NULL CHECK (type IN ('personal', 'guild')),
    points_req INTEGER NOT NULL,
    item_type  INTEGER NOT NULL,
    item_id    INTEGER NOT NULL,
    quantity   INTEGER NOT NULL,
    gr         BOOLEAN NOT NULL DEFAULT false,
    repeatable BOOLEAN NOT NULL DEFAULT false
);

-- Active bead types for the current Diva Defense event.
CREATE TABLE IF NOT EXISTS diva_beads (
    id   SERIAL PRIMARY KEY,
    type INTEGER NOT NULL
);

-- Per-character bead slot assignments with expiry.
CREATE TABLE IF NOT EXISTS diva_beads_assignment (
    id           SERIAL PRIMARY KEY,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    bead_index   INTEGER NOT NULL,
    expiry       TIMESTAMPTZ NOT NULL
);

-- Per-character bead point accumulation log.
CREATE TABLE IF NOT EXISTS diva_beads_points (
    id           SERIAL PRIMARY KEY,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    bead_index   INTEGER NOT NULL,
    points       INTEGER NOT NULL,
    timestamp    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
