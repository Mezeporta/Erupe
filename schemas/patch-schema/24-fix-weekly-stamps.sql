BEGIN;

ALTER TABLE IF EXISTS public.stamps RENAME hl_next TO hl_checked;
ALTER TABLE IF EXISTS public.stamps RENAME ex_next TO ex_checked;

END;
