BEGIN;

ALTER TABLE guilds
    DROP COLUMN rp;

ALTER TABLE guild_characters
    DROP COLUMN is_applicant,
    DROP COLUMN is_sub_leader,
    DROP COLUMN order_index;

END;