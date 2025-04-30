BEGIN;

ALTER TABLE distribution ADD COLUMN rights INTEGER;
ALTER TABLE distribution ADD COLUMN selection BOOLEAN;

END;