-- Add SHA-256 checksum column for savedata integrity verification.
-- Stored as 32 raw bytes (not hex). NULL means no hash computed yet
-- (backwards-compatible with existing data).
ALTER TABLE characters ADD COLUMN IF NOT EXISTS savedata_hash BYTEA;
