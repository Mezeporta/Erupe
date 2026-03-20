BEGIN;

TRUNCATE public.cafebonus;

INSERT INTO public.cafebonus (time_req, item_type, item_id, quantity)
VALUES
    (1800, 17, 0, 50),
    (3600, 17, 0, 100),
    (7200, 17, 0, 200),
    (10800, 17, 0, 300),
    (18000, 17, 0, 350),
    (28800, 17, 0, 500),
    (43200, 17, 0, 500);

END;