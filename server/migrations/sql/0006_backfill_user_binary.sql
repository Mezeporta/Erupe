-- Backfill user_binary rows for characters that were created without one.
-- This fixes #176: new characters could not enter their house.
INSERT INTO user_binary (id)
SELECT c.id FROM characters c
LEFT JOIN user_binary ub ON ub.id = c.id
WHERE ub.id IS NULL;
