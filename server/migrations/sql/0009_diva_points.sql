-- Track per-character Diva Defense (UD) point accumulation per event.
-- Each row records a character's total quest + bonus points for one event.
CREATE TABLE IF NOT EXISTS public.diva_points (
    char_id     integer NOT NULL REFERENCES public.characters(id) ON DELETE CASCADE,
    event_id    integer NOT NULL REFERENCES public.events(id) ON DELETE CASCADE,
    quest_points bigint NOT NULL DEFAULT 0,
    bonus_points bigint NOT NULL DEFAULT 0,
    updated_at  timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (char_id, event_id)
);
