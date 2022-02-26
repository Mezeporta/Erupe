BEGIN;

ALTER TABLE characters
    DROP COLUMN decomyset,
    DROP COLUMN hunternavi,
    DROP COLUMN otomoairou,
    DROP COLUMN partner,
    DROP COLUMN platebox,
    DROP COLUMN platedata,
    DROP COLUMN platemyset,
    DROP COLUMN rengokudata;

END;