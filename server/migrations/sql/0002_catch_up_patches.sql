-- Catch-up migration for databases with partially-applied patch schemas.
--
-- The 0001_init.sql consolidation merged 33 incremental patches (00–32) into one
-- baseline. detectExistingDB marks that baseline as applied for ANY existing database,
-- but users who only ran some of the 33 patches will have schema gaps.
--
-- This migration is:
--   • A no-op on fresh databases (0001 already has everything)
--   • A no-op on fully-patched 9.2 databases
--   • A gap-filler for partially-patched databases
--
-- Omitted patches:
--   15-reset-goocoos   — destructive data reset (NULLs all goocoo columns)
--   20-reset-warehouses — destructive data reset (NULLs all item_box columns)


------------------------------------------------------------------------
-- Patch 00: psn-id (sign_sessions primary key + psn columns)
------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS public.sign_sessions (
    id SERIAL PRIMARY KEY,
    user_id integer,
    char_id integer NOT NULL DEFAULT 0,
    token character varying(16) NOT NULL,
    server_id integer,
    psn_id text
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS psn_id TEXT;

DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'sign_sessions' AND column_name = 'id'
    ) THEN
        ALTER TABLE public.sign_sessions ADD COLUMN id SERIAL;
        ALTER TABLE public.sign_sessions ADD CONSTRAINT sign_sessions_pkey PRIMARY KEY (id);
    END IF;
END $$;

ALTER TABLE public.sign_sessions ALTER COLUMN user_id DROP NOT NULL;

DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'sign_sessions' AND column_name = 'psn_id'
    ) THEN
        ALTER TABLE public.sign_sessions ADD COLUMN psn_id TEXT;
    END IF;
END $$;


------------------------------------------------------------------------
-- Patch 01: wiiu-key
------------------------------------------------------------------------
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS wiiu_key TEXT;


------------------------------------------------------------------------
-- Patch 02: tower
------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS tower (
    char_id INT,
    tr INT,
    trp INT,
    tsp INT,
    block1 INT,
    block2 INT,
    skills TEXT,
    gems TEXT
);

ALTER TABLE IF EXISTS guild_characters ADD COLUMN IF NOT EXISTS tower_mission_1 INT;
ALTER TABLE IF EXISTS guild_characters ADD COLUMN IF NOT EXISTS tower_mission_2 INT;
ALTER TABLE IF EXISTS guild_characters ADD COLUMN IF NOT EXISTS tower_mission_3 INT;
ALTER TABLE IF EXISTS guilds ADD COLUMN IF NOT EXISTS tower_mission_page INT DEFAULT 1;
ALTER TABLE IF EXISTS guilds ADD COLUMN IF NOT EXISTS tower_rp INT DEFAULT 0;


------------------------------------------------------------------------
-- Patch 03: event_quests
------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS event_quests (
    id serial PRIMARY KEY,
    max_players integer,
    quest_type integer NOT NULL,
    quest_id integer NOT NULL,
    mark integer
);

ALTER TABLE IF EXISTS public.servers DROP COLUMN IF EXISTS season;


------------------------------------------------------------------------
-- Patch 04: trend-weapons
------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS public.trend_weapons (
    weapon_id integer NOT NULL,
    weapon_type integer NOT NULL,
    count integer DEFAULT 0,
    PRIMARY KEY (weapon_id)
);


------------------------------------------------------------------------
-- Patch 05: gacha-roll-name
------------------------------------------------------------------------
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'gacha_entries' AND column_name = 'name'
    ) THEN
        ALTER TABLE public.gacha_entries ADD COLUMN name text;
    END IF;
END $$;


------------------------------------------------------------------------
-- Patch 06: goocoo-rename (gook → goocoo)
------------------------------------------------------------------------
DO $$ BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_name = 'gook'
    ) THEN
        ALTER TABLE gook RENAME TO goocoo;
        ALTER TABLE goocoo RENAME COLUMN gook0 TO goocoo0;
        ALTER TABLE goocoo RENAME COLUMN gook1 TO goocoo1;
        ALTER TABLE goocoo RENAME COLUMN gook2 TO goocoo2;
        ALTER TABLE goocoo RENAME COLUMN gook3 TO goocoo3;
        ALTER TABLE goocoo RENAME COLUMN gook4 TO goocoo4;
    END IF;
END $$;


------------------------------------------------------------------------
-- Patch 07: scenarios-counter
------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS scenario_counter (
    id serial PRIMARY KEY,
    scenario_id numeric NOT NULL,
    category_id numeric NOT NULL
);


------------------------------------------------------------------------
-- Patch 08: kill-counts
------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS public.kill_logs (
    id serial PRIMARY KEY,
    character_id integer NOT NULL,
    monster integer NOT NULL,
    quantity integer NOT NULL,
    timestamp timestamp with time zone NOT NULL
);

ALTER TABLE IF EXISTS public.guild_characters
    ADD COLUMN IF NOT EXISTS box_claimed timestamp with time zone DEFAULT now();


------------------------------------------------------------------------
-- Patch 09: fix-guild-treasure
------------------------------------------------------------------------
ALTER TABLE IF EXISTS public.guild_hunts DROP COLUMN IF EXISTS hunters;

ALTER TABLE IF EXISTS public.guild_characters
    ADD COLUMN IF NOT EXISTS treasure_hunt integer;

DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'guild_hunts' AND column_name = 'start'
    ) THEN
        ALTER TABLE public.guild_hunts ADD COLUMN start timestamp with time zone NOT NULL DEFAULT now();
    END IF;
END $$;

ALTER TABLE IF EXISTS public.guild_hunts DROP COLUMN IF EXISTS "return";

DO $$ BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'guild_hunts' AND column_name = 'claimed'
    ) THEN
        ALTER TABLE public.guild_hunts RENAME claimed TO collected;
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS public.guild_hunts_claimed (
    hunt_id integer NOT NULL,
    character_id integer NOT NULL
);

ALTER TABLE IF EXISTS public.guild_hunts DROP COLUMN IF EXISTS treasure;


------------------------------------------------------------------------
-- Patch 10: rework-distributions
------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS public.distribution_items (
    id serial PRIMARY KEY,
    distribution_id integer NOT NULL,
    item_type integer NOT NULL,
    item_id integer,
    quantity integer
);

ALTER TABLE IF EXISTS public.distribution ALTER COLUMN min_hr DROP DEFAULT;
ALTER TABLE IF EXISTS public.distribution ALTER COLUMN max_hr DROP DEFAULT;
ALTER TABLE IF EXISTS public.distribution ALTER COLUMN min_sr DROP DEFAULT;
ALTER TABLE IF EXISTS public.distribution ALTER COLUMN max_sr DROP DEFAULT;
ALTER TABLE IF EXISTS public.distribution ALTER COLUMN min_gr DROP DEFAULT;
ALTER TABLE IF EXISTS public.distribution ALTER COLUMN max_gr DROP DEFAULT;

ALTER TABLE IF EXISTS public.distribution ALTER COLUMN min_hr DROP NOT NULL;
ALTER TABLE IF EXISTS public.distribution ALTER COLUMN max_hr DROP NOT NULL;
ALTER TABLE IF EXISTS public.distribution ALTER COLUMN min_sr DROP NOT NULL;
ALTER TABLE IF EXISTS public.distribution ALTER COLUMN max_sr DROP NOT NULL;
ALTER TABLE IF EXISTS public.distribution ALTER COLUMN min_gr DROP NOT NULL;
ALTER TABLE IF EXISTS public.distribution ALTER COLUMN max_gr DROP NOT NULL;

UPDATE distribution SET min_hr = NULL WHERE min_hr = 65535;
UPDATE distribution SET max_hr = NULL WHERE max_hr = 65535;
UPDATE distribution SET min_sr = NULL WHERE min_sr = 65535;
UPDATE distribution SET max_sr = NULL WHERE max_sr = 65535;
UPDATE distribution SET min_gr = NULL WHERE min_gr = 65535;
UPDATE distribution SET max_gr = NULL WHERE max_gr = 65535;


------------------------------------------------------------------------
-- Patch 11: event-quest-flags
------------------------------------------------------------------------
ALTER TABLE IF EXISTS public.event_quests ADD COLUMN IF NOT EXISTS flags integer;


------------------------------------------------------------------------
-- Patch 12: event_quest_cycling
------------------------------------------------------------------------
ALTER TABLE IF EXISTS public.event_quests
    ADD COLUMN IF NOT EXISTS start_time timestamp with time zone NOT NULL DEFAULT now();

-- Add active_days directly (the original patch added active_duration then renamed it)
ALTER TABLE IF EXISTS public.event_quests ADD COLUMN IF NOT EXISTS active_days int;
ALTER TABLE IF EXISTS public.event_quests ADD COLUMN IF NOT EXISTS inactive_days int;

-- Handle the case where the original patch partially ran (column still named active_duration)
DO $$ BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'event_quests' AND column_name = 'active_duration'
    ) THEN
        ALTER TABLE public.event_quests RENAME active_duration TO active_days;
    END IF;
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'event_quests' AND column_name = 'inactive_duration'
    ) THEN
        ALTER TABLE public.event_quests RENAME inactive_duration TO inactive_days;
    END IF;
END $$;


------------------------------------------------------------------------
-- Patch 13: festa-trial-votes
------------------------------------------------------------------------
ALTER TABLE IF EXISTS public.guild_characters ADD COLUMN IF NOT EXISTS trial_vote integer;


------------------------------------------------------------------------
-- Patch 14: fix-fpoint-trades
------------------------------------------------------------------------
DO $$ BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'fpoint_items' AND column_name = 'item_type'
    ) THEN
        DELETE FROM public.fpoint_items;
        ALTER TABLE public.fpoint_items ALTER COLUMN item_type SET NOT NULL;
        ALTER TABLE public.fpoint_items ALTER COLUMN item_id SET NOT NULL;
        ALTER TABLE public.fpoint_items ALTER COLUMN quantity SET NOT NULL;
        ALTER TABLE public.fpoint_items ALTER COLUMN fpoints SET NOT NULL;
        ALTER TABLE public.fpoint_items DROP COLUMN IF EXISTS trade_type;
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_name = 'fpoint_items' AND column_name = 'buyable'
        ) THEN
            ALTER TABLE public.fpoint_items ADD COLUMN buyable boolean NOT NULL DEFAULT false;
        END IF;
    END IF;
END $$;


------------------------------------------------------------------------
-- Patch 15: reset-goocoos — SKIPPED (destructive data reset)
------------------------------------------------------------------------


------------------------------------------------------------------------
-- Patch 16: discord-password-resets
------------------------------------------------------------------------
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'discord_token'
    ) THEN
        ALTER TABLE public.users ADD COLUMN discord_token text;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'discord_id'
    ) THEN
        ALTER TABLE public.users ADD COLUMN discord_id text;
    END IF;
END $$;


------------------------------------------------------------------------
-- Patch 17: op-accounts
------------------------------------------------------------------------
ALTER TABLE IF EXISTS public.users ADD COLUMN IF NOT EXISTS op boolean;

CREATE TABLE IF NOT EXISTS public.bans (
    user_id integer NOT NULL,
    expires timestamp with time zone,
    PRIMARY KEY (user_id)
);


------------------------------------------------------------------------
-- Patch 18: timer-toggle
------------------------------------------------------------------------
ALTER TABLE users ADD COLUMN IF NOT EXISTS timer bool;


------------------------------------------------------------------------
-- Patch 19: festa-submissions
------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS festa_submissions (
    character_id int NOT NULL,
    guild_id int NOT NULL,
    trial_type int NOT NULL,
    souls int NOT NULL,
    timestamp timestamp with time zone NOT NULL
);

ALTER TABLE guild_characters DROP COLUMN IF EXISTS souls;

DO $$ BEGIN
    ALTER TYPE festival_colour RENAME TO festival_color;
EXCEPTION
    WHEN undefined_object THEN NULL;
    WHEN duplicate_object THEN NULL;
END $$;


------------------------------------------------------------------------
-- Patch 20: reset-warehouses — SKIPPED (destructive data reset)
------------------------------------------------------------------------


------------------------------------------------------------------------
-- Patch 21: rename-hrp (hrp → hr)
------------------------------------------------------------------------
DO $$ BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'characters' AND column_name = 'hrp'
    ) THEN
        ALTER TABLE public.characters RENAME hrp TO hr;
    END IF;
END $$;


------------------------------------------------------------------------
-- Patch 22: clan-changing-room
------------------------------------------------------------------------
ALTER TABLE guilds ADD COLUMN IF NOT EXISTS room_rp INT DEFAULT 0;
ALTER TABLE guilds ADD COLUMN IF NOT EXISTS room_expiry TIMESTAMP WITHOUT TIME ZONE;


------------------------------------------------------------------------
-- Patch 23: rework-distributions-2
------------------------------------------------------------------------
ALTER TABLE IF EXISTS distribution ADD COLUMN IF NOT EXISTS rights INTEGER;
ALTER TABLE IF EXISTS distribution ADD COLUMN IF NOT EXISTS selection BOOLEAN;


------------------------------------------------------------------------
-- Patch 24: fix-weekly-stamps
------------------------------------------------------------------------
DO $$ BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'stamps' AND column_name = 'hl_next'
    ) THEN
        ALTER TABLE public.stamps RENAME hl_next TO hl_checked;
    END IF;
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'stamps' AND column_name = 'ex_next'
    ) THEN
        ALTER TABLE public.stamps RENAME ex_next TO ex_checked;
    END IF;
END $$;


------------------------------------------------------------------------
-- Patch 25: fix-rasta-id
------------------------------------------------------------------------
CREATE SEQUENCE IF NOT EXISTS public.rasta_id_seq;


------------------------------------------------------------------------
-- Patch 26: fix-mail
------------------------------------------------------------------------
ALTER TABLE mail ADD COLUMN IF NOT EXISTS is_sys_message BOOLEAN NOT NULL DEFAULT false;


------------------------------------------------------------------------
-- Patch 27: fix-character-defaults
------------------------------------------------------------------------
UPDATE characters
SET otomoairou = decode(repeat('00', 10), 'hex')
WHERE otomoairou IS NULL OR length(otomoairou) = 0;

UPDATE characters
SET platemyset = decode(repeat('00', 1920), 'hex')
WHERE platemyset IS NULL OR length(platemyset) = 0;


------------------------------------------------------------------------
-- Patch 28: drop-transient-binary-columns
------------------------------------------------------------------------
ALTER TABLE user_binary DROP COLUMN IF EXISTS type2;
ALTER TABLE user_binary DROP COLUMN IF EXISTS type3;
ALTER TABLE characters DROP COLUMN IF EXISTS minidata;


------------------------------------------------------------------------
-- Patch 29: guild-weekly-bonus
------------------------------------------------------------------------
ALTER TABLE IF EXISTS public.guilds
    ADD COLUMN IF NOT EXISTS weekly_bonus_users INT NOT NULL DEFAULT 0;


------------------------------------------------------------------------
-- Patch 30: daily-resets
------------------------------------------------------------------------
ALTER TABLE IF EXISTS public.gacha_stepup
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT now();
ALTER TABLE IF EXISTS public.guilds
    ADD COLUMN IF NOT EXISTS rp_reset_at TIMESTAMP WITH TIME ZONE;


------------------------------------------------------------------------
-- Patch 31: monthly-items
------------------------------------------------------------------------
ALTER TABLE IF EXISTS public.stamps ADD COLUMN IF NOT EXISTS monthly_claimed TIMESTAMP WITH TIME ZONE;
ALTER TABLE IF EXISTS public.stamps ADD COLUMN IF NOT EXISTS monthly_hl_claimed TIMESTAMP WITH TIME ZONE;
ALTER TABLE IF EXISTS public.stamps ADD COLUMN IF NOT EXISTS monthly_ex_claimed TIMESTAMP WITH TIME ZONE;


------------------------------------------------------------------------
-- Patch 32: guild-posts-soft-delete
------------------------------------------------------------------------
ALTER TABLE guild_posts ADD COLUMN IF NOT EXISTS deleted boolean DEFAULT false NOT NULL;
