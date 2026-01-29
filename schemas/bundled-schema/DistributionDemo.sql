BEGIN;

-- Adds a Distribution that can be accepted up to 20 times that gives one of Item Type 30 (Item Box extra page)
INSERT INTO distribution (type, event_name, description, times_acceptable) VALUES (1, 'Extra Item Storage', '~C05Adds one new page to your Item Box.', 20);
INSERT INTO distribution_items (distribution_id, item_type, quantity) VALUES ((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 30, 1);

-- Adds a Distribution that can be accepted up to 20 times that gives one of Item Type 31 (Equipment Box extra page)
INSERT INTO distribution (type, event_name, description, times_acceptable) VALUES (1, 'Extra Equipment Storage', '~C05Adds one new page to your Equipment Box.', 20);
INSERT INTO distribution_items (distribution_id, item_type, quantity) VALUES ((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 31, 1);

END;