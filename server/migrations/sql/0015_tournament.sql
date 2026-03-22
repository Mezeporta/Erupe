BEGIN;

CREATE TABLE IF NOT EXISTS tournaments (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(64) NOT NULL,
    start_time  BIGINT NOT NULL,
    entry_end   BIGINT NOT NULL,
    ranking_end BIGINT NOT NULL,
    reward_end  BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS tournament_cups (
    id            SERIAL PRIMARY KEY,
    tournament_id INTEGER NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    cup_group     SMALLINT NOT NULL,
    cup_type      SMALLINT NOT NULL,
    unk           SMALLINT NOT NULL DEFAULT 0,
    name          VARCHAR(64) NOT NULL,
    description   TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS tournament_sub_events (
    id             SERIAL PRIMARY KEY,
    cup_group      SMALLINT NOT NULL,
    event_sub_type SMALLINT NOT NULL DEFAULT 0,
    quest_file_id  INTEGER NOT NULL DEFAULT 0,
    name           VARCHAR(64) NOT NULL
);

CREATE TABLE IF NOT EXISTS tournament_entries (
    id            SERIAL PRIMARY KEY,
    char_id       INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    tournament_id INTEGER NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (char_id, tournament_id)
);

CREATE TABLE IF NOT EXISTS tournament_results (
    id            SERIAL PRIMARY KEY,
    char_id       INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    tournament_id INTEGER NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    event_id      INTEGER NOT NULL,
    quest_slot    INTEGER NOT NULL DEFAULT 0,
    stage_handle  INTEGER NOT NULL DEFAULT 0,
    submitted_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMIT;
