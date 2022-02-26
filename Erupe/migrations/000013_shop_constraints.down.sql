BEGIN;

ALTER TABLE shop_item_state DROP CONSTRAINT shop_item_state_id_itemhash;
ALTER TABLE shop_item_state ADD CONSTRAINT shop_item_state_itemhash_key UNIQUE (itemhash);

ALTER TABLE stepup_state DROP CONSTRAINT stepup_state_id_shophash;
ALTER TABLE stepup_state ADD CONSTRAINT stepup_state_shophash_key UNIQUE (shophash);

ALTER TABLE lucky_box_state DROP CONSTRAINT lucky_box_state_id_shophash;
ALTER TABLE lucky_box_state ADD CONSTRAINT lucky_box_state_shophash_key UNIQUE (shophash);

END;