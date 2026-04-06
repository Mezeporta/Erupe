-- Diva Defense default prize rewards.
-- Personal track: type='personal', quantity=1 per milestone.
-- Guild track:    type='guild',    quantity=5 per milestone.
-- item_type=26 is Diva Coins; item_id=0 for all.
INSERT INTO diva_prizes (type, points_req, item_type, item_id, quantity, gr, repeatable) VALUES
    ('personal',    500000, 26, 0, 1, false, false),
    ('personal',   1000000, 26, 0, 1, false, false),
    ('personal',   2000000, 26, 0, 1, false, false),
    ('personal',   3000000, 26, 0, 1, false, false),
    ('personal',   5000000, 26, 0, 1, false, false),
    ('personal',   7000000, 26, 0, 1, false, false),
    ('personal',  10000000, 26, 0, 1, false, false),
    ('personal',  15000000, 26, 0, 1, false, false),
    ('personal',  20000000, 26, 0, 1, false, false),
    ('personal',  30000000, 26, 0, 1, false, false),
    ('personal',  50000000, 26, 0, 1, false, false),
    ('personal',  70000000, 26, 0, 1, false, false),
    ('personal', 100000000, 26, 0, 1, false, false),
    ('guild',      500000, 26, 0, 5, false, false),
    ('guild',     1000000, 26, 0, 5, false, false),
    ('guild',     2000000, 26, 0, 5, false, false),
    ('guild',     3000000, 26, 0, 5, false, false),
    ('guild',     5000000, 26, 0, 5, false, false),
    ('guild',     7000000, 26, 0, 5, false, false),
    ('guild',    10000000, 26, 0, 5, false, false),
    ('guild',    15000000, 26, 0, 5, false, false),
    ('guild',    20000000, 26, 0, 5, false, false),
    ('guild',    30000000, 26, 0, 5, false, false),
    ('guild',    50000000, 26, 0, 5, false, false),
    ('guild',    70000000, 26, 0, 5, false, false),
    ('guild',   100000000, 26, 0, 5, false, false)
ON CONFLICT DO NOTHING;
