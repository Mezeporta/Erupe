BEGIN;

ALTER TABLE characters
    ADD COLUMN decomyset bytea,
    ADD COLUMN hunternavi bytea,
    ADD COLUMN otomoairou bytea,
    ADD COLUMN partner bytea,
    ADD COLUMN platebox bytea,
    ADD COLUMN platedata bytea,
    ADD COLUMN platemyset bytea,
    ADD COLUMN trophy bytea,
    ADD COLUMN rengokudata bytea;

END;
