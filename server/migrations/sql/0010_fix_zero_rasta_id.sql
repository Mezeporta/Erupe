-- Heal characters whose rasta_id was clobbered to 0 by the pre-106cf85
-- SaveMercenary bug. A rasta_id of 0 is not a valid sequence value and
-- causes silent save failures on affected characters (see #163).
-- The normal default for characters that never registered a mercenary is NULL.
UPDATE public.characters SET rasta_id = NULL WHERE rasta_id = 0;

ALTER TABLE public.mercenary_logs
    ALTER COLUMN id SET DEFAULT nextval('public.mercenary_logs_id_seq'::regclass);

SELECT setval('public.mercenary_logs_id_seq', COALESCE((SELECT MAX(id) FROM public.mercenary_logs), 0) + 1, false);

BEGIN;

-- strip the first 0x9 (9) bytes, leaving a 0x8b (139) byte payload.
UPDATE characters
SET savemercenary = substring(savemercenary FROM 10)
WHERE octet_length(savemercenary) = 148;

-- Step 2: anything that still isn't exactly 0x8b (139) bytes after that
-- is in an unexpected/corrupt shape — null it out rather than leave
-- a blob the client will misparse.
UPDATE characters
SET savemercenary = NULL
WHERE savemercenary IS NOT NULL
  AND octet_length(savemercenary) <> 139;

COMMIT;