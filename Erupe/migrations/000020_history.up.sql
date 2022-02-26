BEGIN;

CREATE TABLE IF NOT EXISTS public.history
(
    user_id integer,
    admin_id integer,
    report_id integer NOT NULL,
    title text COLLATE pg_catalog."default",
    reason text COLLATE pg_catalog."default",
    CONSTRAINT history_pkey PRIMARY KEY (report_id)
);

END;