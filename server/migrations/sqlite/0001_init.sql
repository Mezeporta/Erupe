-- Erupe consolidated database schema (SQLite)
-- Translated from PostgreSQL 0001_init.sql
-- Compatible with modernc.org/sqlite
--
-- Includes: init.sql (v9.1.0) + 9.2-update.sql + all 33 patch schemas

PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;


--
-- Name: achievements; Type: TABLE
--

CREATE TABLE achievements (
    id integer NOT NULL,
    ach0 integer DEFAULT 0,
    ach1 integer DEFAULT 0,
    ach2 integer DEFAULT 0,
    ach3 integer DEFAULT 0,
    ach4 integer DEFAULT 0,
    ach5 integer DEFAULT 0,
    ach6 integer DEFAULT 0,
    ach7 integer DEFAULT 0,
    ach8 integer DEFAULT 0,
    ach9 integer DEFAULT 0,
    ach10 integer DEFAULT 0,
    ach11 integer DEFAULT 0,
    ach12 integer DEFAULT 0,
    ach13 integer DEFAULT 0,
    ach14 integer DEFAULT 0,
    ach15 integer DEFAULT 0,
    ach16 integer DEFAULT 0,
    ach17 integer DEFAULT 0,
    ach18 integer DEFAULT 0,
    ach19 integer DEFAULT 0,
    ach20 integer DEFAULT 0,
    ach21 integer DEFAULT 0,
    ach22 integer DEFAULT 0,
    ach23 integer DEFAULT 0,
    ach24 integer DEFAULT 0,
    ach25 integer DEFAULT 0,
    ach26 integer DEFAULT 0,
    ach27 integer DEFAULT 0,
    ach28 integer DEFAULT 0,
    ach29 integer DEFAULT 0,
    ach30 integer DEFAULT 0,
    ach31 integer DEFAULT 0,
    ach32 integer DEFAULT 0,
    PRIMARY KEY (id)
);


--
-- Name: bans; Type: TABLE
--

CREATE TABLE bans (
    user_id integer NOT NULL,
    expires TEXT,
    PRIMARY KEY (user_id)
);


--
-- Name: cafe_accepted; Type: TABLE
--

CREATE TABLE cafe_accepted (
    cafe_id integer NOT NULL,
    character_id integer NOT NULL
);


--
-- Name: cafebonus; Type: TABLE
--

CREATE TABLE cafebonus (
    id INTEGER PRIMARY KEY,
    time_req integer NOT NULL,
    item_type integer NOT NULL,
    item_id integer NOT NULL,
    quantity integer NOT NULL
);


--
-- Name: characters; Type: TABLE
--

CREATE TABLE characters (
    id INTEGER PRIMARY KEY,
    user_id bigint,
    is_female boolean,
    is_new_character boolean,
    name TEXT,
    unk_desc_string TEXT,
    gr INTEGER,
    hr INTEGER,
    weapon_type INTEGER,
    last_login integer,
    savedata BLOB,
    decomyset BLOB,
    hunternavi BLOB,
    otomoairou BLOB,
    partner BLOB,
    platebox BLOB,
    platedata BLOB,
    platemyset BLOB,
    rengokudata BLOB,
    savemercenary BLOB,
    restrict_guild_scout boolean DEFAULT false NOT NULL,
    gacha_items BLOB,
    daily_time TEXT,
    house_info BLOB,
    login_boost BLOB,
    skin_hist BLOB,
    kouryou_point integer,
    gcp integer,
    guild_post_checked TEXT DEFAULT CURRENT_TIMESTAMP NOT NULL,
    time_played integer DEFAULT 0 NOT NULL,
    weapon_id integer DEFAULT 0 NOT NULL,
    scenariodata BLOB,
    savefavoritequest BLOB,
    friends text DEFAULT '' NOT NULL,
    blocked text DEFAULT '' NOT NULL,
    deleted boolean DEFAULT false NOT NULL,
    cafe_time integer DEFAULT 0,
    netcafe_points integer DEFAULT 0,
    boost_time TEXT,
    cafe_reset TEXT,
    bonus_quests integer DEFAULT 0 NOT NULL,
    daily_quests integer DEFAULT 0 NOT NULL,
    promo_points integer DEFAULT 0 NOT NULL,
    rasta_id integer,
    pact_id integer,
    stampcard integer DEFAULT 0 NOT NULL,
    mezfes BLOB,
    FOREIGN KEY (user_id) REFERENCES users(id)
);


--
-- Name: distribution; Type: TABLE
--

CREATE TABLE distribution (
    id INTEGER PRIMARY KEY,
    character_id integer,
    type integer NOT NULL,
    deadline TEXT,
    event_name text DEFAULT 'GM Gift!' NOT NULL,
    description text DEFAULT '~C05You received a gift!' NOT NULL,
    times_acceptable integer DEFAULT 1 NOT NULL,
    min_hr integer,
    max_hr integer,
    min_sr integer,
    max_sr integer,
    min_gr integer,
    max_gr integer,
    rights integer,
    selection boolean
);


--
-- Name: distribution_items; Type: TABLE
--

CREATE TABLE distribution_items (
    id INTEGER PRIMARY KEY,
    distribution_id integer NOT NULL,
    item_type integer NOT NULL,
    item_id integer,
    quantity integer
);


--
-- Name: distributions_accepted; Type: TABLE
--

CREATE TABLE distributions_accepted (
    distribution_id integer,
    character_id integer
);


--
-- Name: event_quests; Type: TABLE
--

CREATE TABLE event_quests (
    id INTEGER PRIMARY KEY,
    max_players integer,
    quest_type integer NOT NULL,
    quest_id integer NOT NULL,
    mark integer,
    flags integer,
    start_time TEXT DEFAULT CURRENT_TIMESTAMP NOT NULL,
    active_days integer,
    inactive_days integer
);


--
-- Name: events; Type: TABLE
--

CREATE TABLE events (
    id INTEGER PRIMARY KEY,
    event_type TEXT NOT NULL CHECK (event_type IN ('festa','diva','vs','mezfes')),
    start_time TEXT DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: feature_weapon; Type: TABLE
--

CREATE TABLE feature_weapon (
    start_time TEXT NOT NULL,
    featured integer NOT NULL
);


--
-- Name: festa_prizes; Type: TABLE
--

CREATE TABLE festa_prizes (
    id INTEGER PRIMARY KEY,
    type TEXT NOT NULL CHECK (type IN ('personal','guild')),
    tier integer NOT NULL,
    souls_req integer NOT NULL,
    item_id integer NOT NULL,
    num_item integer NOT NULL
);


--
-- Name: festa_prizes_accepted; Type: TABLE
--

CREATE TABLE festa_prizes_accepted (
    prize_id integer NOT NULL,
    character_id integer NOT NULL
);


--
-- Name: festa_registrations; Type: TABLE
--

CREATE TABLE festa_registrations (
    guild_id integer NOT NULL,
    team TEXT NOT NULL CHECK (team IN ('none','red','blue'))
);


--
-- Name: festa_submissions; Type: TABLE
--

CREATE TABLE festa_submissions (
    character_id integer NOT NULL,
    guild_id integer NOT NULL,
    trial_type integer NOT NULL,
    souls integer NOT NULL,
    "timestamp" TEXT NOT NULL
);


--
-- Name: festa_trials; Type: TABLE
--

CREATE TABLE festa_trials (
    id INTEGER PRIMARY KEY,
    objective integer NOT NULL,
    goal_id integer NOT NULL,
    times_req integer NOT NULL,
    locale_req integer DEFAULT 0 NOT NULL,
    reward integer NOT NULL
);


--
-- Name: fpoint_items; Type: TABLE
--

CREATE TABLE fpoint_items (
    id INTEGER PRIMARY KEY,
    item_type integer NOT NULL,
    item_id integer NOT NULL,
    quantity integer NOT NULL,
    fpoints integer NOT NULL,
    buyable boolean DEFAULT false NOT NULL
);


--
-- Name: gacha_box; Type: TABLE
--

CREATE TABLE gacha_box (
    gacha_id integer,
    entry_id integer,
    character_id integer
);


--
-- Name: gacha_entries; Type: TABLE
--

CREATE TABLE gacha_entries (
    id INTEGER PRIMARY KEY,
    gacha_id integer,
    entry_type integer,
    item_type integer,
    item_number integer,
    item_quantity integer,
    weight integer,
    rarity integer,
    rolls integer,
    frontier_points integer,
    daily_limit integer,
    name text
);


--
-- Name: gacha_items; Type: TABLE
--

CREATE TABLE gacha_items (
    id INTEGER PRIMARY KEY,
    entry_id integer,
    item_type integer,
    item_id integer,
    quantity integer
);


--
-- Name: gacha_shop; Type: TABLE
--

CREATE TABLE gacha_shop (
    id INTEGER PRIMARY KEY,
    min_gr integer,
    min_hr integer,
    name text,
    url_banner text,
    url_feature text,
    url_thumbnail text,
    wide boolean,
    recommended boolean,
    gacha_type integer,
    hidden boolean
);


--
-- Name: gacha_stepup; Type: TABLE
--

CREATE TABLE gacha_stepup (
    gacha_id integer,
    step integer,
    character_id integer,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: goocoo; Type: TABLE
--

CREATE TABLE goocoo (
    id INTEGER PRIMARY KEY,
    goocoo0 BLOB,
    goocoo1 BLOB,
    goocoo2 BLOB,
    goocoo3 BLOB,
    goocoo4 BLOB
);


--
-- Name: guild_adventures; Type: TABLE
--

CREATE TABLE guild_adventures (
    id INTEGER PRIMARY KEY,
    guild_id integer NOT NULL,
    destination integer NOT NULL,
    charge integer DEFAULT 0 NOT NULL,
    depart integer NOT NULL,
    "return" integer NOT NULL,
    collected_by text DEFAULT '' NOT NULL
);


--
-- Name: guild_alliances; Type: TABLE
--

CREATE TABLE guild_alliances (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP NOT NULL,
    parent_id integer NOT NULL,
    sub1_id integer,
    sub2_id integer
);


--
-- Name: guild_applications; Type: TABLE
--

CREATE TABLE guild_applications (
    id INTEGER PRIMARY KEY,
    guild_id integer NOT NULL,
    character_id integer NOT NULL,
    actor_id integer NOT NULL,
    application_type TEXT NOT NULL CHECK (application_type IN ('applied','invited')),
    created_at TEXT DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE (guild_id, character_id),
    FOREIGN KEY (actor_id) REFERENCES characters(id),
    FOREIGN KEY (character_id) REFERENCES characters(id),
    FOREIGN KEY (guild_id) REFERENCES guilds(id)
);


--
-- Name: guild_characters; Type: TABLE
--

CREATE TABLE guild_characters (
    id INTEGER PRIMARY KEY,
    guild_id bigint,
    character_id bigint,
    joined_at TEXT DEFAULT CURRENT_TIMESTAMP,
    avoid_leadership boolean DEFAULT false NOT NULL,
    order_index integer DEFAULT 1 NOT NULL,
    recruiter boolean DEFAULT false NOT NULL,
    rp_today integer DEFAULT 0,
    rp_yesterday integer DEFAULT 0,
    tower_mission_1 integer,
    tower_mission_2 integer,
    tower_mission_3 integer,
    box_claimed TEXT DEFAULT CURRENT_TIMESTAMP,
    treasure_hunt integer,
    trial_vote integer,
    FOREIGN KEY (character_id) REFERENCES characters(id),
    FOREIGN KEY (guild_id) REFERENCES guilds(id)
);


--
-- Name: guild_hunts; Type: TABLE
--

CREATE TABLE guild_hunts (
    id INTEGER PRIMARY KEY,
    guild_id integer NOT NULL,
    host_id integer NOT NULL,
    destination integer NOT NULL,
    level integer NOT NULL,
    acquired boolean DEFAULT false NOT NULL,
    collected boolean DEFAULT false NOT NULL,
    hunt_data BLOB NOT NULL,
    cats_used text NOT NULL,
    start TEXT DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: guild_hunts_claimed; Type: TABLE
--

CREATE TABLE guild_hunts_claimed (
    hunt_id integer NOT NULL,
    character_id integer NOT NULL
);


--
-- Name: guild_meals; Type: TABLE
--

CREATE TABLE guild_meals (
    id INTEGER PRIMARY KEY,
    guild_id integer NOT NULL,
    meal_id integer NOT NULL,
    level integer NOT NULL,
    created_at TEXT
);


--
-- Name: guild_posts; Type: TABLE
--

CREATE TABLE guild_posts (
    id INTEGER PRIMARY KEY,
    guild_id integer NOT NULL,
    author_id integer NOT NULL,
    post_type integer NOT NULL,
    stamp_id integer NOT NULL,
    title text NOT NULL,
    body text NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP NOT NULL,
    liked_by text DEFAULT '' NOT NULL,
    deleted boolean DEFAULT false NOT NULL
);


--
-- Name: guilds; Type: TABLE
--

CREATE TABLE guilds (
    id INTEGER PRIMARY KEY,
    name TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    leader_id integer NOT NULL,
    main_motto integer DEFAULT 0,
    rank_rp integer DEFAULT 0 NOT NULL,
    comment TEXT DEFAULT '' NOT NULL,
    icon BLOB,
    sub_motto integer DEFAULT 0,
    item_box BLOB,
    event_rp integer DEFAULT 0 NOT NULL,
    pugi_name_1 TEXT DEFAULT '',
    pugi_name_2 TEXT DEFAULT '',
    pugi_name_3 TEXT DEFAULT '',
    recruiting boolean DEFAULT true NOT NULL,
    pugi_outfit_1 integer DEFAULT 0 NOT NULL,
    pugi_outfit_2 integer DEFAULT 0 NOT NULL,
    pugi_outfit_3 integer DEFAULT 0 NOT NULL,
    pugi_outfits integer DEFAULT 0 NOT NULL,
    tower_mission_page integer DEFAULT 1,
    tower_rp integer DEFAULT 0,
    room_rp integer DEFAULT 0,
    room_expiry TEXT,
    weekly_bonus_users integer DEFAULT 0 NOT NULL,
    rp_reset_at TEXT
);


--
-- Name: kill_logs; Type: TABLE
--

CREATE TABLE kill_logs (
    id INTEGER PRIMARY KEY,
    character_id integer NOT NULL,
    monster integer NOT NULL,
    quantity integer NOT NULL,
    "timestamp" TEXT NOT NULL
);


--
-- Name: login_boost; Type: TABLE
--

CREATE TABLE login_boost (
    char_id integer,
    week_req integer,
    expiration TEXT,
    reset TEXT
);


--
-- Name: mail; Type: TABLE
--

CREATE TABLE mail (
    id INTEGER PRIMARY KEY,
    sender_id integer NOT NULL,
    recipient_id integer NOT NULL,
    subject TEXT DEFAULT '' NOT NULL,
    body TEXT DEFAULT '' NOT NULL,
    read boolean DEFAULT false NOT NULL,
    attached_item_received boolean DEFAULT false NOT NULL,
    attached_item integer,
    attached_item_amount integer DEFAULT 1 NOT NULL,
    is_guild_invite boolean DEFAULT false NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted boolean DEFAULT false NOT NULL,
    locked boolean DEFAULT false NOT NULL,
    is_sys_message boolean DEFAULT false NOT NULL,
    FOREIGN KEY (recipient_id) REFERENCES characters(id),
    FOREIGN KEY (sender_id) REFERENCES characters(id)
);


--
-- Name: rengoku_score; Type: TABLE
--

CREATE TABLE rengoku_score (
    character_id integer NOT NULL,
    max_stages_mp integer,
    max_points_mp integer,
    max_stages_sp integer,
    max_points_sp integer,
    PRIMARY KEY (character_id)
);


--
-- Name: scenario_counter; Type: TABLE
--

CREATE TABLE scenario_counter (
    id INTEGER PRIMARY KEY,
    scenario_id numeric NOT NULL,
    category_id numeric NOT NULL
);


--
-- Name: servers; Type: TABLE
--

CREATE TABLE servers (
    server_id integer NOT NULL,
    current_players integer NOT NULL,
    world_name text,
    world_description text,
    land integer
);


--
-- Name: shop_items; Type: TABLE
--

CREATE TABLE shop_items (
    shop_type integer,
    shop_id integer,
    id INTEGER PRIMARY KEY,
    item_id INTEGER,
    cost integer,
    quantity INTEGER,
    min_hr INTEGER,
    min_sr INTEGER,
    min_gr INTEGER,
    store_level INTEGER,
    max_quantity INTEGER,
    road_floors INTEGER,
    road_fatalis INTEGER
);


--
-- Name: shop_items_bought; Type: TABLE
--

CREATE TABLE shop_items_bought (
    character_id integer,
    shop_item_id integer,
    bought integer
);

CREATE UNIQUE INDEX IF NOT EXISTS shop_items_bought_character_item_unique
    ON shop_items_bought (character_id, shop_item_id);


--
-- Name: sign_sessions; Type: TABLE
--

CREATE TABLE sign_sessions (
    user_id integer,
    char_id integer,
    token TEXT NOT NULL,
    server_id integer,
    id INTEGER PRIMARY KEY,
    psn_id text
);


--
-- Name: stamps; Type: TABLE
--

CREATE TABLE stamps (
    character_id integer NOT NULL,
    hl_total integer DEFAULT 0,
    hl_redeemed integer DEFAULT 0,
    hl_checked TEXT,
    ex_total integer DEFAULT 0,
    ex_redeemed integer DEFAULT 0,
    ex_checked TEXT,
    monthly_claimed TEXT,
    monthly_hl_claimed TEXT,
    monthly_ex_claimed TEXT,
    PRIMARY KEY (character_id)
);


--
-- Name: titles; Type: TABLE
--

CREATE TABLE titles (
    id integer NOT NULL,
    char_id integer NOT NULL,
    unlocked_at TEXT,
    updated_at TEXT
);


--
-- Name: tower; Type: TABLE
--

CREATE TABLE tower (
    char_id integer,
    tr integer,
    trp integer,
    tsp integer,
    block1 integer,
    block2 integer,
    skills text,
    gems text
);


--
-- Name: trend_weapons; Type: TABLE
--

CREATE TABLE trend_weapons (
    weapon_id integer NOT NULL,
    weapon_type integer NOT NULL,
    count integer DEFAULT 0,
    PRIMARY KEY (weapon_id)
);


--
-- Name: user_binary; Type: TABLE
--

CREATE TABLE user_binary (
    id INTEGER PRIMARY KEY,
    house_tier BLOB,
    house_state integer,
    house_password text,
    house_data BLOB,
    house_furniture BLOB,
    bookshelf BLOB,
    gallery BLOB,
    tore BLOB,
    garden BLOB,
    mission BLOB
);


--
-- Name: users; Type: TABLE
--

CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    username text NOT NULL UNIQUE,
    password text NOT NULL,
    item_box BLOB,
    rights integer DEFAULT 12 NOT NULL,
    last_character integer DEFAULT 0,
    last_login TEXT,
    return_expires TEXT,
    gacha_premium integer,
    gacha_trial integer,
    frontier_points integer,
    psn_id text,
    wiiu_key text,
    discord_token text,
    discord_id text,
    op boolean,
    timer boolean
);


--
-- Name: warehouse; Type: TABLE
--

CREATE TABLE warehouse (
    character_id integer NOT NULL,
    item0 BLOB,
    item1 BLOB,
    item2 BLOB,
    item3 BLOB,
    item4 BLOB,
    item5 BLOB,
    item6 BLOB,
    item7 BLOB,
    item8 BLOB,
    item9 BLOB,
    item10 BLOB,
    item0name text,
    item1name text,
    item2name text,
    item3name text,
    item4name text,
    item5name text,
    item6name text,
    item7name text,
    item8name text,
    item9name text,
    equip0 BLOB,
    equip1 BLOB,
    equip2 BLOB,
    equip3 BLOB,
    equip4 BLOB,
    equip5 BLOB,
    equip6 BLOB,
    equip7 BLOB,
    equip8 BLOB,
    equip9 BLOB,
    equip10 BLOB,
    equip0name text,
    equip1name text,
    equip2name text,
    equip3name text,
    equip4name text,
    equip5name text,
    equip6name text,
    equip7name text,
    equip8name text,
    equip9name text,
    PRIMARY KEY (character_id)
);


--
-- Indexes
--

CREATE INDEX guild_application_type_index ON guild_applications (application_type);

CREATE UNIQUE INDEX guild_character_unique_index ON guild_characters (character_id);

CREATE INDEX mail_recipient_deleted_created_id_index ON mail (recipient_id, deleted, created_at DESC, id DESC);
