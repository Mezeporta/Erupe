BEGIN;

ALTER TABLE characters
    DROP COLUMN minidata,
    DROP COLUMN gacha_trial,
    DROP COLUMN gacha_prem,
    DROP COLUMN gacha_items,
    DROP COLUMN daily_time,
    DROP COLUMN frontier_points,
	DROP COLUMN netcafe_points,
	DROP COLUMN house_info,
	DROP COLUMN login_boost,
	DROP COLUMN skin_hist,
	DROP COLUMN gcp;

DROP TABLE fpoint_items;
DROP TABLE gacha_shop;
DROP TABLE gacha_shop_items;
DROP TABLE lucky_box_state;
DROP TABLE stepup_state;
DROP TABLE normal_shop_items;
DROP TABLE shop_item_state;

END;