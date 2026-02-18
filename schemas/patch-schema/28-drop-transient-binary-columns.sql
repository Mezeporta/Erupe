-- Drop transient binary columns that are now memory-only.
-- UserBinary type2/type3 and characters.minidata are session state
-- resent by the client on every login; they do not need persistence.

ALTER TABLE user_binary DROP COLUMN IF EXISTS type2;
ALTER TABLE user_binary DROP COLUMN IF EXISTS type3;
ALTER TABLE characters DROP COLUMN IF EXISTS minidata;
