BEGIN;

ALTER TABLE guild_characters
    ADD COLUMN is_applicant  bool NOT NULL DEFAULT false,
    ADD COLUMN is_sub_leader bool NOT NULL DEFAULT false,
    ADD COLUMN order_index   int NOT NULL DEFAULT 1;

ALTER TABLE guilds
    ADD COLUMN rp uint16 NOT NULL DEFAULT 0;

END;
