BEGIN;
CREATE TABLE mail
(
    id                     SERIAL    NOT NULL PRIMARY KEY,
    sender_id              INT       NOT NULL REFERENCES characters (id),
    recipient_id           INT       NOT NULL REFERENCES characters (id),
    subject                VARCHAR   NOT NULL DEFAULT '',
    body                   VARCHAR   NOT NULL DEFAULT '',
    read                   BOOL      NOT NULL DEFAULT FALSE,
    attached_item_received BOOL      NOT NULL DEFAULT FALSE,
    attached_item          INT                DEFAULT NULL,
    attached_item_amount   INT       NOT NULL DEFAULT 1,
    is_guild_invite        BOOL      NOT NULL DEFAULT FALSE,
    created_at             TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted                BOOL      NOT NULL DEFAULT FALSE
);

CREATE INDEX mail_recipient_deleted_created_id_index ON mail (recipient_id, deleted, created_at DESC, id DESC);
END;