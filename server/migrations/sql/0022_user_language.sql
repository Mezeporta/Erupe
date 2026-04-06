-- Per-user preferred language for server-generated content (chat commands,
-- mail templates, future localized quest/scenario text). NULL means "use the
-- server default" (config.Language). See #188 (server-side multi-language
-- support), phase A (plumbing).

ALTER TABLE public.users
    ADD COLUMN IF NOT EXISTS language TEXT;
