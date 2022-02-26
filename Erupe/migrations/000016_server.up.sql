BEGIN;
-- Table: public.servers

-- DROP TABLE IF EXISTS public.servers;

CREATE TABLE IF NOT EXISTS public.servers
(
    server_id integer NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1 ),
    server_name text COLLATE pg_catalog."default",
    season integer,
    current_players integer,
    event_id integer,
    event_expiration integer,
    CONSTRAINT servers_pkey PRIMARY KEY (server_id)
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.servers
    OWNER to postgres;
END;