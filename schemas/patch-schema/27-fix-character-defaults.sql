BEGIN;

-- Initialize otomoairou (mercenary data) with default empty data for characters that have NULL or empty values
-- This prevents error logs when loading mercenary data during zone transitions
UPDATE characters
SET otomoairou = decode(repeat('00', 10), 'hex')
WHERE otomoairou IS NULL OR length(otomoairou) = 0;

-- Initialize platemyset (plate configuration) with default empty data for characters that have NULL or empty values
-- This prevents error logs when loading plate data during zone transitions
UPDATE characters
SET platemyset = decode(repeat('00', 1920), 'hex')
WHERE platemyset IS NULL OR length(platemyset) = 0;

COMMIT;
