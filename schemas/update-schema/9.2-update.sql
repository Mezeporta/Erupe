BEGIN;

DROP TABLE IF EXISTS public.fpoint_items;

CREATE TABLE IF NOT EXISTS public.fpoint_items (
    id serial PRIMARY KEY,
    item_type integer,
    item_id integer,
    quantity integer,
    fpoints integer,
    trade_type integer
);

ALTER TABLE IF EXISTS public.characters ADD bonus_quests INT NOT NULL DEFAULT 0;

ALTER TABLE IF EXISTS public.characters ADD daily_quests INT NOT NULL DEFAULT 0;

ALTER TABLE IF EXISTS public.characters ADD promo_points INT NOT NULL DEFAULT 0;

ALTER TABLE IF EXISTS public.guild_characters ADD rp_today INT DEFAULT 0;

ALTER TABLE IF EXISTS public.guild_characters ADD rp_yesterday INT DEFAULT 0;

UPDATE public.characters SET savemercenary = NULL;

ALTER TABLE IF EXISTS public.characters ADD rasta_id INT;

ALTER TABLE IF EXISTS public.characters ADD pact_id INT;

ALTER TABLE IF EXISTS public.characters ADD stampcard INT NOT NULL DEFAULT 0;

ALTER TABLE IF EXISTS public.characters DROP COLUMN IF EXISTS gacha_prem;

ALTER TABLE IF EXISTS public.characters DROP COLUMN IF EXISTS gacha_trial;

ALTER TABLE IF EXISTS public.characters DROP COLUMN IF EXISTS frontier_points;

ALTER TABLE IF EXISTS public.users ADD IF NOT EXISTS gacha_premium INT;

ALTER TABLE IF EXISTS public.users ADD IF NOT EXISTS gacha_trial INT;

ALTER TABLE IF EXISTS public.users ADD IF NOT EXISTS frontier_points INT;

DROP TABLE IF EXISTS public.gacha_shop;

CREATE TABLE IF NOT EXISTS public.gacha_shop (
    id SERIAL PRIMARY KEY,
    min_gr INTEGER,
    min_hr INTEGER,
    name TEXT,
    url_banner TEXT,
    url_feature TEXT,
    url_thumbnail TEXT,
    wide BOOLEAN,
    recommended BOOLEAN,
    gacha_type INTEGER,
    hidden BOOLEAN
);

DROP TABLE IF EXISTS public.gacha_shop_items;

CREATE TABLE IF NOT EXISTS public.gacha_entries (
    id            SERIAL PRIMARY KEY,
    gacha_id      INTEGER,
    entry_type    INTEGER,
    item_type     INTEGER,
    item_number   INTEGER,
    item_quantity INTEGER,
    weight        INTEGER,
    rarity        INTEGER,
    rolls         INTEGER,
    frontier_points INTEGER,
    daily_limit   INTEGER
);

CREATE TABLE IF NOT EXISTS public.gacha_items (
    id        SERIAL PRIMARY KEY,
    entry_id  INTEGER,
    item_type INTEGER,
    item_id   INTEGER,
    quantity  INTEGER
);

DROP TABLE IF EXISTS public.stepup_state;

CREATE TABLE IF NOT EXISTS public.gacha_stepup (
    gacha_id INTEGER,
    step INTEGER,
    character_id INTEGER
);

DROP TABLE IF EXISTS public.lucky_box_state;

CREATE TABLE IF NOT EXISTS public.gacha_box (
    gacha_id INTEGER,
    entry_id INTEGER,
    character_id INTEGER
);

DROP TABLE IF EXISTS public.login_boost_state;

CREATE TABLE IF NOT EXISTS public.login_boost (
    char_id INTEGER,
    week_req INTEGER,
    expiration TIMESTAMP WITH TIME ZONE,
    reset TIMESTAMP WITH TIME ZONE
);

ALTER TABLE IF EXISTS public.characters ADD COLUMN mezfes BYTEA;

ALTER TABLE IF EXISTS public.characters ALTER COLUMN daily_time TYPE TIMESTAMP WITH TIME ZONE;

ALTER TABLE IF EXISTS public.characters ALTER COLUMN guild_post_checked TYPE TIMESTAMP WITH TIME ZONE;

ALTER TABLE IF EXISTS public.characters ALTER COLUMN boost_time TYPE TIMESTAMP WITH TIME ZONE;

ALTER TABLE IF EXISTS public.characters ADD COLUMN IF NOT EXISTS cafe_reset TIMESTAMP WITHOUT TIME ZONE;

ALTER TABLE IF EXISTS public.characters ALTER COLUMN cafe_reset TYPE TIMESTAMP WITH TIME ZONE;

ALTER TABLE IF EXISTS public.distribution ALTER COLUMN deadline TYPE TIMESTAMP WITH TIME ZONE;

ALTER TABLE IF EXISTS public.events ALTER COLUMN start_time TYPE TIMESTAMP WITH TIME ZONE;

ALTER TABLE IF EXISTS public.feature_weapon ALTER COLUMN start_time TYPE TIMESTAMP WITH TIME ZONE;

CREATE TABLE IF NOT EXISTS public.feature_weapon
(
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    featured INTEGER NOT NULL
);

ALTER TABLE IF EXISTS public.guild_alliances ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE;

ALTER TABLE IF EXISTS public.guild_applications ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE;

ALTER TABLE IF EXISTS public.guild_characters ALTER COLUMN joined_at TYPE TIMESTAMP WITH TIME ZONE;

ALTER TABLE IF EXISTS public.guild_posts ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE;

ALTER TABLE IF EXISTS public.characters ALTER COLUMN daily_time TYPE TIMESTAMP WITH TIME ZONE;

ALTER TABLE IF EXISTS public.guilds ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE;

ALTER TABLE IF EXISTS public.mail ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE;

ALTER TABLE IF EXISTS public.stamps ALTER COLUMN hl_next TYPE TIMESTAMP WITH TIME ZONE;

ALTER TABLE IF EXISTS public.stamps ALTER COLUMN ex_next TYPE TIMESTAMP WITH TIME ZONE;

ALTER TABLE IF EXISTS public.titles ALTER COLUMN unlocked_at TYPE TIMESTAMP WITH TIME ZONE;

ALTER TABLE IF EXISTS public.titles ALTER COLUMN updated_at TYPE TIMESTAMP WITH TIME ZONE;

ALTER TABLE IF EXISTS public.users ALTER COLUMN last_login TYPE TIMESTAMP WITH TIME ZONE;

ALTER TABLE IF EXISTS public.users ALTER COLUMN return_expires TYPE TIMESTAMP WITH TIME ZONE;

ALTER TABLE IF EXISTS public.guild_meals DROP COLUMN IF EXISTS expires;

ALTER TABLE IF EXISTS public.guild_meals ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE;

DROP TABLE IF EXISTS public.account_ban;

DROP TABLE IF EXISTS public.account_history;

DROP TABLE IF EXISTS public.account_moderation;

DROP TABLE IF EXISTS public.account_sub;

DROP TABLE IF EXISTS public.history;

DROP TABLE IF EXISTS public.questlists;

DROP TABLE IF EXISTS public.schema_migrations;

DROP TABLE IF EXISTS public.user_binaries;

DROP PROCEDURE IF EXISTS raviinit;

DROP PROCEDURE IF EXISTS ravireset;

ALTER TABLE IF EXISTS public.normal_shop_items RENAME TO shop_items;

ALTER TABLE IF EXISTS public.shop_items RENAME COLUMN shoptype TO shop_type;

ALTER TABLE IF EXISTS public.shop_items RENAME COLUMN shopid TO shop_id;

ALTER TABLE IF EXISTS public.shop_items RENAME COLUMN itemhash TO id;

ALTER TABLE IF EXISTS public.shop_items RENAME COLUMN itemid TO item_id;

ALTER TABLE IF EXISTS public.shop_items ALTER COLUMN points TYPE integer;

ALTER TABLE IF EXISTS public.shop_items RENAME COLUMN points TO cost;

ALTER TABLE IF EXISTS public.shop_items RENAME COLUMN tradequantity TO quantity;

ALTER TABLE IF EXISTS public.shop_items RENAME COLUMN rankreqlow TO min_hr;

ALTER TABLE IF EXISTS public.shop_items RENAME COLUMN rankreqhigh TO min_sr;

ALTER TABLE IF EXISTS public.shop_items RENAME COLUMN rankreqg TO min_gr;

ALTER TABLE IF EXISTS public.shop_items RENAME COLUMN storelevelreq TO store_level;

ALTER TABLE IF EXISTS public.shop_items RENAME COLUMN maximumquantity TO max_quantity;

ALTER TABLE IF EXISTS public.shop_items DROP COLUMN IF EXISTS boughtquantity;

ALTER TABLE IF EXISTS public.shop_items RENAME COLUMN roadfloorsrequired TO road_floors;

ALTER TABLE IF EXISTS public.shop_items RENAME COLUMN weeklyfataliskills TO road_fatalis;

ALTER TABLE public.shop_items RENAME CONSTRAINT normal_shop_items_pkey TO shop_items_pkey;

ALTER TABLE IF EXISTS public.shop_items DROP CONSTRAINT IF EXISTS normal_shop_items_itemhash_key;

CREATE SEQUENCE IF NOT EXISTS public.shop_items_id_seq;

ALTER SEQUENCE IF EXISTS public.shop_items_id_seq OWNER TO postgres;

ALTER TABLE IF EXISTS public.shop_items ALTER COLUMN id SET DEFAULT nextval('shop_items_id_seq'::regclass);

ALTER SEQUENCE IF EXISTS public.shop_items_id_seq OWNED BY shop_items.id;

SELECT setval('shop_items_id_seq', (SELECT MAX(id) FROM public.shop_items));

DROP TABLE IF EXISTS public.shop_item_state;

CREATE TABLE IF NOT EXISTS public.shop_items_bought (
    character_id INTEGER,
    shop_item_id INTEGER,
    bought INTEGER
);

UPDATE users SET rights = rights-2;

ALTER TABLE IF EXISTS public.users ALTER COLUMN rights SET DEFAULT 12;

END;