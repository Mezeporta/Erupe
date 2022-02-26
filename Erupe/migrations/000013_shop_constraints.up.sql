BEGIN;

ALTER TABLE shop_item_state DROP CONSTRAINT shop_item_state_itemhash_key;
ALTER TABLE shop_item_state ADD CONSTRAINT shop_item_state_id_itemhash UNIQUE(char_id, itemhash);

ALTER TABLE stepup_state DROP CONSTRAINT stepup_state_shophash_key;
ALTER TABLE stepup_state ADD CONSTRAINT stepup_state_id_shophash UNIQUE(char_id, shophash);

ALTER TABLE lucky_box_state DROP CONSTRAINT lucky_box_state_shophash_key;
ALTER TABLE lucky_box_state ADD CONSTRAINT lucky_box_state_id_shophash UNIQUE(char_id, shophash);

END;