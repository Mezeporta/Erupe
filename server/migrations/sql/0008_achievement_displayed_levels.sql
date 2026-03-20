-- 0008: Add displayed_levels to achievements table for rank-up notifications (#165).
-- Stores 33 bytes (one level per achievement) representing the last level
-- the client acknowledged seeing. NULL means never displayed (shows all rank-ups).
ALTER TABLE public.achievements ADD COLUMN IF NOT EXISTS displayed_levels bytea;
