BEGIN;

CREATE TABLE IF NOT EXISTS public.gook
(
    id serial NOT NULL PRIMARY KEY,
    gook0 bytea,
    gook1 bytea,
    gook2 bytea,
    gook3 bytea,
    gook4 bytea,
    gook5 bytea,
    gook0status boolean,
    gook1status boolean,
    gook2status boolean,
    gook3status boolean,
    gook4status boolean,
    gook5status boolean
);

END;