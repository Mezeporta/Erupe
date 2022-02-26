BEGIN;
ALTER TABLE characters
    ADD COLUMN minidata bytea,
    ADD COLUMN gacha_trial int,
    ADD COLUMN gacha_prem int,
    ADD COLUMN gacha_items bytea,
    ADD COLUMN daily_time timestamp,
    ADD COLUMN frontier_points int,
	ADD COLUMN netcafe_points int,
	ADD COLUMN house_info bytea,
	ADD COLUMN login_boost bytea,
	ADD COLUMN skin_hist bytea,
	ADD COLUMN kouryou_point int,
	ADD COLUMN gcp int;

CREATE TABLE fpoint_items
(
	hash int,
	itemType uint8,
	itemID uint16,
	quant uint16,
	itemValue uint16,
	tradeType uint8
);


CREATE TABLE gacha_shop
(
	hash bigint,
	reqGR int,
	reqHR int,
	gachaName varchar(255),
	gachaLink0 varchar(255),
	gachaLink1 varchar(255),
	gachaLink2 varchar(255),
	extraIcon int,
	gachaType int,
	hideFlag bool
);

CREATE TABLE gacha_shop_items
(
	shophash int,
	entryType uint8,
	itemhash int UNIQUE NOT NULL,
	currType uint8,
	currNumber uint16,
	currQuant uint16,
	percentage uint16,
	rarityIcon uint8,
	rollsCount uint8,
	itemCount uint8,
	dailyLimit uint8,
	itemType int[],
	itemId int[],
	quantity int[]
);

CREATE TABLE lucky_box_state
(
    char_id bigint REFERENCES characters (id),
	shophash int UNIQUE NOT NULL,
	used_itemhash int[]
);


CREATE TABLE stepup_state
(
    char_id bigint REFERENCES characters (id),
	shophash int UNIQUE NOT NULL,
	step_progression int,
    step_time timestamp
);

CREATE TABLE normal_shop_items
(
	shoptype int,
	shopid int,
	itemhash int UNIQUE NOT NULL,
	itemID uint16,
	Points uint16,
	TradeQuantity uint16,
	rankReqLow uint16,
	rankReqHigh uint16,
	rankReqG uint16,
	storeLevelReq uint16,
	maximumQuantity uint16,
	boughtQuantity uint16,
	roadFloorsRequired uint16,
	weeklyFatalisKills uint16
);

CREATE TABLE shop_item_state
(
    char_id bigint REFERENCES characters (id),
	itemhash int UNIQUE NOT NULL,
	usedquantity int
);

END;