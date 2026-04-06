-- Heal characters whose rasta_id was clobbered to 0 by the pre-106cf85
-- SaveMercenary bug. A rasta_id of 0 is not a valid sequence value and
-- causes silent save failures on affected characters (see #163).
-- The normal default for characters that never registered a mercenary is NULL.
UPDATE public.characters SET rasta_id = NULL WHERE rasta_id = 0;
