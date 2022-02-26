BEGIN;

CREATE TABLE IF NOT EXISTS public.event_week
(
    id integer NOT NULL,
    event_id integer NOT NULL,
    date_expiration integer NOT NULL,
    CONSTRAINT event_week_pkey PRIMARY KEY (id)
);

END;