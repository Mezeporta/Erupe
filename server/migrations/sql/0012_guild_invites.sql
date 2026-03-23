-- Dedicated table for guild-initiated scout invitations, separate from
-- player-initiated applications. This gives each invitation a real serial PK
-- so the client's InvitationID field can map to an actual database row
-- instead of being aliased to the character ID.
CREATE TABLE guild_invites (
    id           serial PRIMARY KEY,
    guild_id     integer REFERENCES guilds(id),
    character_id integer REFERENCES characters(id),
    actor_id     integer REFERENCES characters(id),
    created_at   timestamptz NOT NULL DEFAULT now()
);

-- Migrate any existing scout invitations from guild_applications.
INSERT INTO guild_invites (guild_id, character_id, actor_id, created_at)
SELECT guild_id, character_id, actor_id, COALESCE(created_at, now())
FROM guild_applications
WHERE application_type = 'invited';

DELETE FROM guild_applications WHERE application_type = 'invited';
