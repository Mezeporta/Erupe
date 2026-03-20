-- Erupe consolidated database schema
-- This file is auto-generated. Do not edit manually.
-- To update, modify future migration files (0002_*.sql, etc.)
--
-- Includes: init.sql (v9.1.0) + 9.2-update.sql + all 33 patch schemas


--
-- Name: event_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.event_type AS ENUM (
    'festa',
    'diva',
    'vs',
    'mezfes'
);


--
-- Name: festival_color; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.festival_color AS ENUM (
    'none',
    'red',
    'blue'
);


--
-- Name: guild_application_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.guild_application_type AS ENUM (
    'applied',
    'invited'
);


--
-- Name: prize_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.prize_type AS ENUM (
    'personal',
    'guild'
);


--
-- Name: uint16; Type: DOMAIN; Schema: public; Owner: -
--

CREATE DOMAIN public.uint16 AS integer
	CONSTRAINT uint16_check CHECK (((VALUE >= 0) AND (VALUE <= 65536)));


--
-- Name: uint8; Type: DOMAIN; Schema: public; Owner: -
--

CREATE DOMAIN public.uint8 AS smallint
	CONSTRAINT uint8_check CHECK (((VALUE >= 0) AND (VALUE <= 255)));


--
-- Name: achievements; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.achievements (
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
    ach32 integer DEFAULT 0
);


--
-- Name: airou_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.airou_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: bans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.bans (
    user_id integer NOT NULL,
    expires timestamp with time zone
);


--
-- Name: cafe_accepted; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cafe_accepted (
    cafe_id integer NOT NULL,
    character_id integer NOT NULL
);


--
-- Name: cafebonus; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cafebonus (
    id integer NOT NULL,
    time_req integer NOT NULL,
    item_type integer NOT NULL,
    item_id integer NOT NULL,
    quantity integer NOT NULL
);


--
-- Name: cafebonus_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.cafebonus_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: cafebonus_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.cafebonus_id_seq OWNED BY public.cafebonus.id;


--
-- Name: characters; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.characters (
    id integer NOT NULL,
    user_id bigint,
    is_female boolean,
    is_new_character boolean,
    name character varying(15),
    unk_desc_string character varying(31),
    gr public.uint16,
    hr public.uint16,
    weapon_type public.uint16,
    last_login integer,
    savedata bytea,
    decomyset bytea,
    hunternavi bytea,
    otomoairou bytea,
    partner bytea,
    platebox bytea,
    platedata bytea,
    platemyset bytea,
    rengokudata bytea,
    savemercenary bytea,
    restrict_guild_scout boolean DEFAULT false NOT NULL,
    gacha_items bytea,
    daily_time timestamp with time zone,
    house_info bytea,
    login_boost bytea,
    skin_hist bytea,
    kouryou_point integer,
    gcp integer,
    guild_post_checked timestamp with time zone DEFAULT now() NOT NULL,
    time_played integer DEFAULT 0 NOT NULL,
    weapon_id integer DEFAULT 0 NOT NULL,
    scenariodata bytea,
    savefavoritequest bytea,
    friends text DEFAULT ''::text NOT NULL,
    blocked text DEFAULT ''::text NOT NULL,
    deleted boolean DEFAULT false NOT NULL,
    cafe_time integer DEFAULT 0,
    netcafe_points integer DEFAULT 0,
    boost_time timestamp with time zone,
    cafe_reset timestamp with time zone,
    bonus_quests integer DEFAULT 0 NOT NULL,
    daily_quests integer DEFAULT 0 NOT NULL,
    promo_points integer DEFAULT 0 NOT NULL,
    rasta_id integer,
    pact_id integer,
    stampcard integer DEFAULT 0 NOT NULL,
    mezfes bytea
);


--
-- Name: characters_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.characters_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: characters_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.characters_id_seq OWNED BY public.characters.id;


--
-- Name: distribution; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.distribution (
    id integer NOT NULL,
    character_id integer,
    type integer NOT NULL,
    deadline timestamp with time zone,
    event_name text DEFAULT 'GM Gift!'::text NOT NULL,
    description text DEFAULT '~C05You received a gift!'::text NOT NULL,
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
-- Name: distribution_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.distribution_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: distribution_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.distribution_id_seq OWNED BY public.distribution.id;


--
-- Name: distribution_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.distribution_items (
    id integer NOT NULL,
    distribution_id integer NOT NULL,
    item_type integer NOT NULL,
    item_id integer,
    quantity integer
);


--
-- Name: distribution_items_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.distribution_items_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: distribution_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.distribution_items_id_seq OWNED BY public.distribution_items.id;


--
-- Name: distributions_accepted; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.distributions_accepted (
    distribution_id integer,
    character_id integer
);


--
-- Name: event_quests; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.event_quests (
    id integer NOT NULL,
    max_players integer,
    quest_type integer NOT NULL,
    quest_id integer NOT NULL,
    mark integer,
    flags integer,
    start_time timestamp with time zone DEFAULT now() NOT NULL,
    active_days integer,
    inactive_days integer
);


--
-- Name: event_quests_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.event_quests_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: event_quests_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.event_quests_id_seq OWNED BY public.event_quests.id;


--
-- Name: events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.events (
    id integer NOT NULL,
    event_type public.event_type NOT NULL,
    start_time timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: events_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.events_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.events_id_seq OWNED BY public.events.id;


--
-- Name: feature_weapon; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.feature_weapon (
    start_time timestamp with time zone NOT NULL,
    featured integer NOT NULL
);


--
-- Name: festa_prizes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.festa_prizes (
    id integer NOT NULL,
    type public.prize_type NOT NULL,
    tier integer NOT NULL,
    souls_req integer NOT NULL,
    item_id integer NOT NULL,
    num_item integer NOT NULL
);


--
-- Name: festa_prizes_accepted; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.festa_prizes_accepted (
    prize_id integer NOT NULL,
    character_id integer NOT NULL
);


--
-- Name: festa_prizes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.festa_prizes_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: festa_prizes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.festa_prizes_id_seq OWNED BY public.festa_prizes.id;


--
-- Name: festa_registrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.festa_registrations (
    guild_id integer NOT NULL,
    team public.festival_color NOT NULL
);


--
-- Name: festa_submissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.festa_submissions (
    character_id integer NOT NULL,
    guild_id integer NOT NULL,
    trial_type integer NOT NULL,
    souls integer NOT NULL,
    "timestamp" timestamp with time zone NOT NULL
);


--
-- Name: festa_trials; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.festa_trials (
    id integer NOT NULL,
    objective integer NOT NULL,
    goal_id integer NOT NULL,
    times_req integer NOT NULL,
    locale_req integer DEFAULT 0 NOT NULL,
    reward integer NOT NULL
);


--
-- Name: festa_trials_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.festa_trials_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: festa_trials_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.festa_trials_id_seq OWNED BY public.festa_trials.id;


--
-- Name: fpoint_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.fpoint_items (
    id integer NOT NULL,
    item_type integer NOT NULL,
    item_id integer NOT NULL,
    quantity integer NOT NULL,
    fpoints integer NOT NULL,
    buyable boolean DEFAULT false NOT NULL
);


--
-- Name: fpoint_items_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.fpoint_items_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: fpoint_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.fpoint_items_id_seq OWNED BY public.fpoint_items.id;


--
-- Name: gacha_box; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gacha_box (
    gacha_id integer,
    entry_id integer,
    character_id integer
);


--
-- Name: gacha_entries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gacha_entries (
    id integer NOT NULL,
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
-- Name: gacha_entries_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.gacha_entries_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: gacha_entries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.gacha_entries_id_seq OWNED BY public.gacha_entries.id;


--
-- Name: gacha_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gacha_items (
    id integer NOT NULL,
    entry_id integer,
    item_type integer,
    item_id integer,
    quantity integer
);


--
-- Name: gacha_items_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.gacha_items_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: gacha_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.gacha_items_id_seq OWNED BY public.gacha_items.id;


--
-- Name: gacha_shop; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gacha_shop (
    id integer NOT NULL,
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
-- Name: gacha_shop_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.gacha_shop_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: gacha_shop_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.gacha_shop_id_seq OWNED BY public.gacha_shop.id;


--
-- Name: gacha_stepup; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gacha_stepup (
    gacha_id integer,
    step integer,
    character_id integer,
    created_at timestamp with time zone DEFAULT now()
);


--
-- Name: goocoo; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.goocoo (
    id integer CONSTRAINT gook_id_not_null NOT NULL,
    goocoo0 bytea,
    goocoo1 bytea,
    goocoo2 bytea,
    goocoo3 bytea,
    goocoo4 bytea
);


--
-- Name: gook_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.gook_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: gook_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.gook_id_seq OWNED BY public.goocoo.id;


--
-- Name: guild_adventures; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.guild_adventures (
    id integer NOT NULL,
    guild_id integer NOT NULL,
    destination integer NOT NULL,
    charge integer DEFAULT 0 NOT NULL,
    depart integer NOT NULL,
    return integer NOT NULL,
    collected_by text DEFAULT ''::text NOT NULL
);


--
-- Name: guild_adventures_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.guild_adventures_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: guild_adventures_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.guild_adventures_id_seq OWNED BY public.guild_adventures.id;


--
-- Name: guild_alliances; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.guild_alliances (
    id integer NOT NULL,
    name character varying(24) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    parent_id integer NOT NULL,
    sub1_id integer,
    sub2_id integer
);


--
-- Name: guild_alliances_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.guild_alliances_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: guild_alliances_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.guild_alliances_id_seq OWNED BY public.guild_alliances.id;


--
-- Name: guild_applications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.guild_applications (
    id integer NOT NULL,
    guild_id integer NOT NULL,
    character_id integer NOT NULL,
    actor_id integer NOT NULL,
    application_type public.guild_application_type NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: guild_applications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.guild_applications_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: guild_applications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.guild_applications_id_seq OWNED BY public.guild_applications.id;


--
-- Name: guild_characters; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.guild_characters (
    id integer NOT NULL,
    guild_id bigint,
    character_id bigint,
    joined_at timestamp with time zone DEFAULT now(),
    avoid_leadership boolean DEFAULT false NOT NULL,
    order_index integer DEFAULT 1 NOT NULL,
    recruiter boolean DEFAULT false NOT NULL,
    rp_today integer DEFAULT 0,
    rp_yesterday integer DEFAULT 0,
    tower_mission_1 integer,
    tower_mission_2 integer,
    tower_mission_3 integer,
    box_claimed timestamp with time zone DEFAULT now(),
    treasure_hunt integer,
    trial_vote integer
);


--
-- Name: guild_characters_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.guild_characters_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: guild_characters_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.guild_characters_id_seq OWNED BY public.guild_characters.id;


--
-- Name: guild_hunts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.guild_hunts (
    id integer NOT NULL,
    guild_id integer NOT NULL,
    host_id integer NOT NULL,
    destination integer NOT NULL,
    level integer NOT NULL,
    acquired boolean DEFAULT false NOT NULL,
    collected boolean DEFAULT false CONSTRAINT guild_hunts_claimed_not_null NOT NULL,
    hunt_data bytea NOT NULL,
    cats_used text NOT NULL,
    start timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: guild_hunts_claimed; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.guild_hunts_claimed (
    hunt_id integer NOT NULL,
    character_id integer NOT NULL
);


--
-- Name: guild_hunts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.guild_hunts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: guild_hunts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.guild_hunts_id_seq OWNED BY public.guild_hunts.id;


--
-- Name: guild_meals; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.guild_meals (
    id integer NOT NULL,
    guild_id integer NOT NULL,
    meal_id integer NOT NULL,
    level integer NOT NULL,
    created_at timestamp with time zone
);


--
-- Name: guild_meals_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.guild_meals_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: guild_meals_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.guild_meals_id_seq OWNED BY public.guild_meals.id;


--
-- Name: guild_posts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.guild_posts (
    id integer NOT NULL,
    guild_id integer NOT NULL,
    author_id integer NOT NULL,
    post_type integer NOT NULL,
    stamp_id integer NOT NULL,
    title text NOT NULL,
    body text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    liked_by text DEFAULT ''::text NOT NULL,
    deleted boolean DEFAULT false NOT NULL
);


--
-- Name: guild_posts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.guild_posts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: guild_posts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.guild_posts_id_seq OWNED BY public.guild_posts.id;


--
-- Name: guilds; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.guilds (
    id integer NOT NULL,
    name character varying(24),
    created_at timestamp with time zone DEFAULT now(),
    leader_id integer NOT NULL,
    main_motto integer DEFAULT 0,
    rank_rp integer DEFAULT 0 NOT NULL,
    comment character varying(255) DEFAULT ''::character varying NOT NULL,
    icon bytea,
    sub_motto integer DEFAULT 0,
    item_box bytea,
    event_rp integer DEFAULT 0 NOT NULL,
    pugi_name_1 character varying(12) DEFAULT ''::character varying,
    pugi_name_2 character varying(12) DEFAULT ''::character varying,
    pugi_name_3 character varying(12) DEFAULT ''::character varying,
    recruiting boolean DEFAULT true NOT NULL,
    pugi_outfit_1 integer DEFAULT 0 NOT NULL,
    pugi_outfit_2 integer DEFAULT 0 NOT NULL,
    pugi_outfit_3 integer DEFAULT 0 NOT NULL,
    pugi_outfits integer DEFAULT 0 NOT NULL,
    tower_mission_page integer DEFAULT 1,
    tower_rp integer DEFAULT 0,
    room_rp integer DEFAULT 0,
    room_expiry timestamp without time zone,
    weekly_bonus_users integer DEFAULT 0 NOT NULL,
    rp_reset_at timestamp with time zone
);


--
-- Name: guilds_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.guilds_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: guilds_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.guilds_id_seq OWNED BY public.guilds.id;


--
-- Name: kill_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.kill_logs (
    id integer NOT NULL,
    character_id integer NOT NULL,
    monster integer NOT NULL,
    quantity integer NOT NULL,
    "timestamp" timestamp with time zone NOT NULL
);


--
-- Name: kill_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.kill_logs_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: kill_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.kill_logs_id_seq OWNED BY public.kill_logs.id;


--
-- Name: login_boost; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.login_boost (
    char_id integer,
    week_req integer,
    expiration timestamp with time zone,
    reset timestamp with time zone
);


--
-- Name: mail; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.mail (
    id integer NOT NULL,
    sender_id integer NOT NULL,
    recipient_id integer NOT NULL,
    subject character varying DEFAULT ''::character varying NOT NULL,
    body character varying DEFAULT ''::character varying NOT NULL,
    read boolean DEFAULT false NOT NULL,
    attached_item_received boolean DEFAULT false NOT NULL,
    attached_item integer,
    attached_item_amount integer DEFAULT 1 NOT NULL,
    is_guild_invite boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted boolean DEFAULT false NOT NULL,
    locked boolean DEFAULT false NOT NULL,
    is_sys_message boolean DEFAULT false NOT NULL
);


--
-- Name: mail_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.mail_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: mail_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.mail_id_seq OWNED BY public.mail.id;


--
-- Name: rasta_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.rasta_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: rengoku_score; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.rengoku_score (
    character_id integer NOT NULL,
    max_stages_mp integer,
    max_points_mp integer,
    max_stages_sp integer,
    max_points_sp integer
);


--
-- Name: scenario_counter; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.scenario_counter (
    id integer NOT NULL,
    scenario_id numeric NOT NULL,
    category_id numeric NOT NULL
);


--
-- Name: scenario_counter_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.scenario_counter_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: scenario_counter_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.scenario_counter_id_seq OWNED BY public.scenario_counter.id;


--
-- Name: servers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.servers (
    server_id integer NOT NULL,
    current_players integer NOT NULL,
    world_name text,
    world_description text,
    land integer
);


--
-- Name: shop_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.shop_items (
    shop_type integer,
    shop_id integer,
    id integer CONSTRAINT normal_shop_items_itemhash_not_null NOT NULL,
    item_id public.uint16,
    cost integer,
    quantity public.uint16,
    min_hr public.uint16,
    min_sr public.uint16,
    min_gr public.uint16,
    store_level public.uint16,
    max_quantity public.uint16,
    road_floors public.uint16,
    road_fatalis public.uint16
);


--
-- Name: shop_items_bought; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.shop_items_bought (
    character_id integer,
    shop_item_id integer,
    bought integer
);

CREATE UNIQUE INDEX IF NOT EXISTS shop_items_bought_character_item_unique
    ON public.shop_items_bought (character_id, shop_item_id);


--
-- Name: shop_items_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.shop_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: shop_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.shop_items_id_seq OWNED BY public.shop_items.id;


--
-- Name: sign_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sign_sessions (
    user_id integer,
    char_id integer,
    token character varying(16) NOT NULL,
    server_id integer,
    id integer NOT NULL,
    psn_id text
);


--
-- Name: sign_sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sign_sessions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sign_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sign_sessions_id_seq OWNED BY public.sign_sessions.id;


--
-- Name: stamps; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.stamps (
    character_id integer NOT NULL,
    hl_total integer DEFAULT 0,
    hl_redeemed integer DEFAULT 0,
    hl_checked timestamp with time zone,
    ex_total integer DEFAULT 0,
    ex_redeemed integer DEFAULT 0,
    ex_checked timestamp with time zone,
    monthly_claimed timestamp with time zone,
    monthly_hl_claimed timestamp with time zone,
    monthly_ex_claimed timestamp with time zone
);


--
-- Name: titles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.titles (
    id integer NOT NULL,
    char_id integer NOT NULL,
    unlocked_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: tower; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tower (
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
-- Name: trend_weapons; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.trend_weapons (
    weapon_id integer NOT NULL,
    weapon_type integer NOT NULL,
    count integer DEFAULT 0
);


--
-- Name: user_binary; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_binary (
    id integer NOT NULL,
    house_tier bytea,
    house_state integer,
    house_password text,
    house_data bytea,
    house_furniture bytea,
    bookshelf bytea,
    gallery bytea,
    tore bytea,
    garden bytea,
    mission bytea
);


--
-- Name: user_binary_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_binary_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_binary_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_binary_id_seq OWNED BY public.user_binary.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id integer NOT NULL,
    username text NOT NULL,
    password text NOT NULL,
    item_box bytea,
    rights integer DEFAULT 12 NOT NULL,
    last_character integer DEFAULT 0,
    last_login timestamp with time zone,
    return_expires timestamp with time zone,
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
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: warehouse; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.warehouse (
    character_id integer NOT NULL,
    item0 bytea,
    item1 bytea,
    item2 bytea,
    item3 bytea,
    item4 bytea,
    item5 bytea,
    item6 bytea,
    item7 bytea,
    item8 bytea,
    item9 bytea,
    item10 bytea,
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
    equip0 bytea,
    equip1 bytea,
    equip2 bytea,
    equip3 bytea,
    equip4 bytea,
    equip5 bytea,
    equip6 bytea,
    equip7 bytea,
    equip8 bytea,
    equip9 bytea,
    equip10 bytea,
    equip0name text,
    equip1name text,
    equip2name text,
    equip3name text,
    equip4name text,
    equip5name text,
    equip6name text,
    equip7name text,
    equip8name text,
    equip9name text
);


--
-- Name: cafebonus id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cafebonus ALTER COLUMN id SET DEFAULT nextval('public.cafebonus_id_seq'::regclass);


--
-- Name: characters id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.characters ALTER COLUMN id SET DEFAULT nextval('public.characters_id_seq'::regclass);


--
-- Name: distribution id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.distribution ALTER COLUMN id SET DEFAULT nextval('public.distribution_id_seq'::regclass);


--
-- Name: distribution_items id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.distribution_items ALTER COLUMN id SET DEFAULT nextval('public.distribution_items_id_seq'::regclass);


--
-- Name: event_quests id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_quests ALTER COLUMN id SET DEFAULT nextval('public.event_quests_id_seq'::regclass);


--
-- Name: events id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events ALTER COLUMN id SET DEFAULT nextval('public.events_id_seq'::regclass);


--
-- Name: festa_prizes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.festa_prizes ALTER COLUMN id SET DEFAULT nextval('public.festa_prizes_id_seq'::regclass);


--
-- Name: festa_trials id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.festa_trials ALTER COLUMN id SET DEFAULT nextval('public.festa_trials_id_seq'::regclass);


--
-- Name: fpoint_items id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fpoint_items ALTER COLUMN id SET DEFAULT nextval('public.fpoint_items_id_seq'::regclass);


--
-- Name: gacha_entries id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gacha_entries ALTER COLUMN id SET DEFAULT nextval('public.gacha_entries_id_seq'::regclass);


--
-- Name: gacha_items id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gacha_items ALTER COLUMN id SET DEFAULT nextval('public.gacha_items_id_seq'::regclass);


--
-- Name: gacha_shop id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gacha_shop ALTER COLUMN id SET DEFAULT nextval('public.gacha_shop_id_seq'::regclass);


--
-- Name: goocoo id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.goocoo ALTER COLUMN id SET DEFAULT nextval('public.gook_id_seq'::regclass);


--
-- Name: guild_adventures id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_adventures ALTER COLUMN id SET DEFAULT nextval('public.guild_adventures_id_seq'::regclass);


--
-- Name: guild_alliances id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_alliances ALTER COLUMN id SET DEFAULT nextval('public.guild_alliances_id_seq'::regclass);


--
-- Name: guild_applications id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_applications ALTER COLUMN id SET DEFAULT nextval('public.guild_applications_id_seq'::regclass);


--
-- Name: guild_characters id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_characters ALTER COLUMN id SET DEFAULT nextval('public.guild_characters_id_seq'::regclass);


--
-- Name: guild_hunts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_hunts ALTER COLUMN id SET DEFAULT nextval('public.guild_hunts_id_seq'::regclass);


--
-- Name: guild_meals id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_meals ALTER COLUMN id SET DEFAULT nextval('public.guild_meals_id_seq'::regclass);


--
-- Name: guild_posts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_posts ALTER COLUMN id SET DEFAULT nextval('public.guild_posts_id_seq'::regclass);


--
-- Name: guilds id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guilds ALTER COLUMN id SET DEFAULT nextval('public.guilds_id_seq'::regclass);


--
-- Name: kill_logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.kill_logs ALTER COLUMN id SET DEFAULT nextval('public.kill_logs_id_seq'::regclass);


--
-- Name: mail id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mail ALTER COLUMN id SET DEFAULT nextval('public.mail_id_seq'::regclass);


--
-- Name: scenario_counter id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.scenario_counter ALTER COLUMN id SET DEFAULT nextval('public.scenario_counter_id_seq'::regclass);


--
-- Name: shop_items id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.shop_items ALTER COLUMN id SET DEFAULT nextval('public.shop_items_id_seq'::regclass);


--
-- Name: sign_sessions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sign_sessions ALTER COLUMN id SET DEFAULT nextval('public.sign_sessions_id_seq'::regclass);


--
-- Name: user_binary id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_binary ALTER COLUMN id SET DEFAULT nextval('public.user_binary_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: achievements achievements_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.achievements
    ADD CONSTRAINT achievements_pkey PRIMARY KEY (id);


--
-- Name: bans bans_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bans
    ADD CONSTRAINT bans_pkey PRIMARY KEY (user_id);


--
-- Name: cafebonus cafebonus_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cafebonus
    ADD CONSTRAINT cafebonus_pkey PRIMARY KEY (id);


--
-- Name: characters characters_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.characters
    ADD CONSTRAINT characters_pkey PRIMARY KEY (id);


--
-- Name: distribution_items distribution_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.distribution_items
    ADD CONSTRAINT distribution_items_pkey PRIMARY KEY (id);


--
-- Name: distribution distribution_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.distribution
    ADD CONSTRAINT distribution_pkey PRIMARY KEY (id);


--
-- Name: event_quests event_quests_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_quests
    ADD CONSTRAINT event_quests_pkey PRIMARY KEY (id);


--
-- Name: events events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_pkey PRIMARY KEY (id);


--
-- Name: festa_prizes festa_prizes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.festa_prizes
    ADD CONSTRAINT festa_prizes_pkey PRIMARY KEY (id);


--
-- Name: festa_trials festa_trials_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.festa_trials
    ADD CONSTRAINT festa_trials_pkey PRIMARY KEY (id);


--
-- Name: fpoint_items fpoint_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fpoint_items
    ADD CONSTRAINT fpoint_items_pkey PRIMARY KEY (id);


--
-- Name: gacha_entries gacha_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gacha_entries
    ADD CONSTRAINT gacha_entries_pkey PRIMARY KEY (id);


--
-- Name: gacha_items gacha_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gacha_items
    ADD CONSTRAINT gacha_items_pkey PRIMARY KEY (id);


--
-- Name: gacha_shop gacha_shop_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gacha_shop
    ADD CONSTRAINT gacha_shop_pkey PRIMARY KEY (id);


--
-- Name: goocoo gook_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.goocoo
    ADD CONSTRAINT gook_pkey PRIMARY KEY (id);


--
-- Name: guild_adventures guild_adventures_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_adventures
    ADD CONSTRAINT guild_adventures_pkey PRIMARY KEY (id);


--
-- Name: guild_alliances guild_alliances_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_alliances
    ADD CONSTRAINT guild_alliances_pkey PRIMARY KEY (id);


--
-- Name: guild_applications guild_application_character_id; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_applications
    ADD CONSTRAINT guild_application_character_id UNIQUE (guild_id, character_id);


--
-- Name: guild_applications guild_applications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_applications
    ADD CONSTRAINT guild_applications_pkey PRIMARY KEY (id);


--
-- Name: guild_characters guild_characters_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_characters
    ADD CONSTRAINT guild_characters_pkey PRIMARY KEY (id);


--
-- Name: guild_hunts guild_hunts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_hunts
    ADD CONSTRAINT guild_hunts_pkey PRIMARY KEY (id);


--
-- Name: guild_meals guild_meals_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_meals
    ADD CONSTRAINT guild_meals_pkey PRIMARY KEY (id);


--
-- Name: guild_posts guild_posts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_posts
    ADD CONSTRAINT guild_posts_pkey PRIMARY KEY (id);


--
-- Name: guilds guilds_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guilds
    ADD CONSTRAINT guilds_pkey PRIMARY KEY (id);


--
-- Name: kill_logs kill_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.kill_logs
    ADD CONSTRAINT kill_logs_pkey PRIMARY KEY (id);


--
-- Name: mail mail_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mail
    ADD CONSTRAINT mail_pkey PRIMARY KEY (id);


--
-- Name: rengoku_score rengoku_score_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rengoku_score
    ADD CONSTRAINT rengoku_score_pkey PRIMARY KEY (character_id);


--
-- Name: scenario_counter scenario_counter_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.scenario_counter
    ADD CONSTRAINT scenario_counter_pkey PRIMARY KEY (id);


--
-- Name: shop_items shop_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.shop_items
    ADD CONSTRAINT shop_items_pkey PRIMARY KEY (id);


--
-- Name: sign_sessions sign_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sign_sessions
    ADD CONSTRAINT sign_sessions_pkey PRIMARY KEY (id);


--
-- Name: stamps stamps_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stamps
    ADD CONSTRAINT stamps_pkey PRIMARY KEY (character_id);


--
-- Name: trend_weapons trend_weapons_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trend_weapons
    ADD CONSTRAINT trend_weapons_pkey PRIMARY KEY (weapon_id);


--
-- Name: user_binary user_binary_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_binary
    ADD CONSTRAINT user_binary_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: warehouse warehouse_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.warehouse
    ADD CONSTRAINT warehouse_pkey PRIMARY KEY (character_id);


--
-- Name: guild_application_type_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX guild_application_type_index ON public.guild_applications USING btree (application_type);


--
-- Name: guild_character_unique_index; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX guild_character_unique_index ON public.guild_characters USING btree (character_id);


--
-- Name: mail_recipient_deleted_created_id_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX mail_recipient_deleted_created_id_index ON public.mail USING btree (recipient_id, deleted, created_at DESC, id DESC);


--
-- Name: characters characters_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.characters
    ADD CONSTRAINT characters_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: guild_applications guild_applications_actor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_applications
    ADD CONSTRAINT guild_applications_actor_id_fkey FOREIGN KEY (actor_id) REFERENCES public.characters(id);


--
-- Name: guild_applications guild_applications_character_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_applications
    ADD CONSTRAINT guild_applications_character_id_fkey FOREIGN KEY (character_id) REFERENCES public.characters(id);


--
-- Name: guild_applications guild_applications_guild_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_applications
    ADD CONSTRAINT guild_applications_guild_id_fkey FOREIGN KEY (guild_id) REFERENCES public.guilds(id);


--
-- Name: guild_characters guild_characters_character_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_characters
    ADD CONSTRAINT guild_characters_character_id_fkey FOREIGN KEY (character_id) REFERENCES public.characters(id);


--
-- Name: guild_characters guild_characters_guild_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_characters
    ADD CONSTRAINT guild_characters_guild_id_fkey FOREIGN KEY (guild_id) REFERENCES public.guilds(id);


--
-- Name: mail mail_recipient_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mail
    ADD CONSTRAINT mail_recipient_id_fkey FOREIGN KEY (recipient_id) REFERENCES public.characters(id);


--
-- Name: mail mail_sender_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mail
    ADD CONSTRAINT mail_sender_id_fkey FOREIGN KEY (sender_id) REFERENCES public.characters(id);
