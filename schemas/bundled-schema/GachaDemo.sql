BEGIN;

-- Start Normal Demo
INSERT INTO gacha_shop (min_gr, min_hr, name, url_banner, url_feature, url_thumbnail, wide, recommended, gacha_type, hidden)
    VALUES (0, 0, 'Normal Demo',
    'http://img4.imagetitan.com/img4/QeRWNAviFD8UoTx/26/26_template_innerbanner.png',
    'http://img4.imagetitan.com/img4/QeRWNAviFD8UoTx/26/26_template_feature.png',
    'http://img4.imagetitan.com/img4/small/26/26_template_outerbanner.png',
    false, false, 0, false);

-- Create two different 'rolls', the first rolls once for 1z, the second rolls eleven times for 10z
INSERT INTO gacha_entries (gacha_id, entry_type, item_type, item_number, item_quantity, weight, rarity, rolls, daily_limit, frontier_points)
VALUES
    ((SELECT id FROM gacha_shop ORDER BY id DESC LIMIT 1), 0, 10, 1, 0, 0, 0, 1, 0, 0),
    ((SELECT id FROM gacha_shop ORDER BY id DESC LIMIT 1), 1, 10, 10, 0, 0, 0, 11, 0, 0);

-- Creates a prize of 1z with a weighted chance of 100
INSERT INTO gacha_entries (gacha_id, entry_type, item_type, item_number, item_quantity, weight, rarity, rolls, daily_limit, frontier_points)
    VALUES ((SELECT id FROM gacha_shop ORDER BY id DESC LIMIT 1), 100, 0, 0, 0, 100, 0, 0, 0, 0);
INSERT INTO gacha_items (entry_id, item_type, item_id, quantity)
    VALUES ((SELECT id FROM gacha_entries ORDER BY id DESC LIMIT 1), 10, 1, 0);

-- Creates a prize of 2z with a weighted chance of 70
INSERT INTO gacha_entries (gacha_id, entry_type, item_type, item_number, item_quantity, weight, rarity, rolls, daily_limit, frontier_points)
    VALUES ((SELECT id FROM gacha_shop ORDER BY id DESC LIMIT 1), 100, 0, 0, 0, 70, 1, 0, 0, 0);
INSERT INTO gacha_items (entry_id, item_type, item_id, quantity)
    VALUES ((SELECT id FROM gacha_entries ORDER BY id DESC LIMIT 1), 10, 2, 0);

-- Creates a prize of 3z with a weighted chance of 10
INSERT INTO gacha_entries (gacha_id, entry_type, item_type, item_number, item_quantity, weight, rarity, rolls, daily_limit, frontier_points)
    VALUES ((SELECT id FROM gacha_shop ORDER BY id DESC LIMIT 1), 100, 0, 0, 0, 10, 2, 0, 0, 0);
INSERT INTO gacha_items (entry_id, item_type, item_id, quantity)
    VALUES ((SELECT id FROM gacha_entries ORDER BY id DESC LIMIT 1), 10, 3, 0);
-- End Normal Demo

-- Start Step-Up Demo
INSERT INTO gacha_shop (min_gr, min_hr, name, url_banner, url_feature, url_thumbnail, wide, recommended, gacha_type, hidden)
VALUES (0, 0, 'Step-Up Demo', '', '', '', false, false, 1, false);

-- Create two 'steps', the first costs 1z, the second costs 2z
-- The first step has zero rolls so it will only give the prizes directly linked to the entry ID, being 1z
INSERT INTO gacha_entries (gacha_id, entry_type, item_type, item_number, item_quantity, weight, rarity, rolls, daily_limit, frontier_points)
    VALUES ((SELECT id FROM gacha_shop ORDER BY id DESC LIMIT 1), 0, 10, 1, 0, 0, 0, 0, 0, 0);
INSERT INTO gacha_items (entry_id, item_type, item_id, quantity)
    VALUES ((SELECT id FROM gacha_entries ORDER BY id DESC LIMIT 1), 10, 1, 0);

-- The second step has one roll on the random prize list as will as the direct prize, being 3z
INSERT INTO gacha_entries (gacha_id, entry_type, item_type, item_number, item_quantity, weight, rarity, rolls, daily_limit, frontier_points)
    VALUES ((SELECT id FROM gacha_shop ORDER BY id DESC LIMIT 1), 1, 10, 2, 0, 0, 0, 1, 0, 0);
INSERT INTO gacha_items (entry_id, item_type, item_id, quantity)
    VALUES ((SELECT id FROM gacha_entries ORDER BY id DESC LIMIT 1), 10, 3, 0);

-- Set up two random prizes, the first gives 1z, the second gives 2z
INSERT INTO gacha_entries (gacha_id, entry_type, item_type, item_number, item_quantity, weight, rarity, rolls, daily_limit, frontier_points)
    VALUES ((SELECT id FROM gacha_shop ORDER BY id DESC LIMIT 1), 100, 0, 0, 0, 100, 0, 0, 0, 0);
INSERT INTO gacha_items (entry_id, item_type, item_id, quantity)
    VALUES ((SELECT id FROM gacha_entries ORDER BY id DESC LIMIT 1), 10, 1, 0);

INSERT INTO gacha_entries (gacha_id, entry_type, item_type, item_number, item_quantity, weight, rarity, rolls, daily_limit, frontier_points)
    VALUES ((SELECT id FROM gacha_shop ORDER BY id DESC LIMIT 1), 100, 0, 0, 0, 90, 1, 0, 0, 0);
INSERT INTO gacha_items (entry_id, item_type, item_id, quantity)
    VALUES ((SELECT id FROM gacha_entries ORDER BY id DESC LIMIT 1), 10, 2, 0);
-- End Step-Up Demo

-- Start Box Demo
INSERT INTO gacha_shop (min_gr, min_hr, name, url_banner, url_feature, url_thumbnail, wide, recommended, gacha_type, hidden)
VALUES (0, 0, 'Box Demo', '', '', '', false, false, 4, false);

-- Create two different 'rolls', the first rolls once for 1z, the second rolls twice for 2z
INSERT INTO gacha_entries (gacha_id, entry_type, item_type, item_number, item_quantity, weight, rarity, rolls, daily_limit, frontier_points)
VALUES
    ((SELECT id FROM gacha_shop ORDER BY id DESC LIMIT 1), 0, 10, 1, 0, 0, 0, 1, 0, 0),
    ((SELECT id FROM gacha_shop ORDER BY id DESC LIMIT 1), 1, 10, 2, 0, 0, 0, 2, 0, 0);

-- Create five different 'Box' items, weight is always 0 for these
INSERT INTO gacha_entries (gacha_id, entry_type, item_type, item_number, item_quantity, weight, rarity, rolls, daily_limit, frontier_points)
    VALUES ((SELECT id FROM gacha_shop ORDER BY id DESC LIMIT 1), 100, 0, 0, 0, 0, 0, 0, 0, 0);
INSERT INTO gacha_items (entry_id, item_type, item_id, quantity)
    VALUES ((SELECT id FROM gacha_entries ORDER BY id DESC LIMIT 1), 10, 1, 0);

INSERT INTO gacha_entries (gacha_id, entry_type, item_type, item_number, item_quantity, weight, rarity, rolls, daily_limit, frontier_points)
    VALUES ((SELECT id FROM gacha_shop ORDER BY id DESC LIMIT 1), 100, 0, 0, 0, 0, 0, 0, 0, 0);
INSERT INTO gacha_items (entry_id, item_type, item_id, quantity)
    VALUES ((SELECT id FROM gacha_entries ORDER BY id DESC LIMIT 1), 10, 1, 0);

INSERT INTO gacha_entries (gacha_id, entry_type, item_type, item_number, item_quantity, weight, rarity, rolls, daily_limit, frontier_points)
    VALUES ((SELECT id FROM gacha_shop ORDER BY id DESC LIMIT 1), 100, 0, 0, 0, 0, 0, 0, 0, 0);
INSERT INTO gacha_items (entry_id, item_type, item_id, quantity)
    VALUES ((SELECT id FROM gacha_entries ORDER BY id DESC LIMIT 1), 10, 1, 0);

INSERT INTO gacha_entries (gacha_id, entry_type, item_type, item_number, item_quantity, weight, rarity, rolls, daily_limit, frontier_points)
    VALUES ((SELECT id FROM gacha_shop ORDER BY id DESC LIMIT 1), 100, 0, 0, 0, 0, 0, 0, 0, 0);
INSERT INTO gacha_items (entry_id, item_type, item_id, quantity)
    VALUES ((SELECT id FROM gacha_entries ORDER BY id DESC LIMIT 1), 10, 2, 0);

INSERT INTO gacha_entries (gacha_id, entry_type, item_type, item_number, item_quantity, weight, rarity, rolls, daily_limit, frontier_points)
    VALUES ((SELECT id FROM gacha_shop ORDER BY id DESC LIMIT 1), 100, 0, 0, 0, 0, 0, 0, 0, 0);
INSERT INTO gacha_items (entry_id, item_type, item_id, quantity)
    VALUES ((SELECT id FROM gacha_entries ORDER BY id DESC LIMIT 1), 10, 3, 0);
-- End Box Demo

END;