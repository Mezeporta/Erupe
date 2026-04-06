-- Tournament #150 default data.
-- One tournament is inserted that starts immediately and has a wide window so operators
-- can adjust the timestamps after installation. The sub-events and cups are seeded
-- idempotently via ON CONFLICT DO NOTHING.
-- Cup groups: 16 = speed hunt (Brachydios variants), 17 = guild hunt, 6 = fishing size.
-- Cup types: 7 = speed hunt, 6 = fishing size.

BEGIN;

-- Default tournament (always active on a fresh install).
-- start_time = now, entry_end = +3 days, ranking_end = +13 days, reward_end = +20 days.
INSERT INTO tournaments (name, start_time, entry_end, ranking_end, reward_end)
SELECT
    'Tournament #150',
    EXTRACT(epoch FROM NOW())::bigint,
    EXTRACT(epoch FROM NOW() + INTERVAL '3 days')::bigint,
    EXTRACT(epoch FROM NOW() + INTERVAL '13 days')::bigint,
    EXTRACT(epoch FROM NOW() + INTERVAL '20 days')::bigint
WHERE NOT EXISTS (SELECT 1 FROM tournaments);

-- Sub-events (shared across tournaments; NOT tournament-specific).
-- CupGroup 16: Speed hunt Brachydios variants (event_sub_type 0-14, quest_file_id 60691).
INSERT INTO tournament_sub_events (cup_group, event_sub_type, quest_file_id, name)
SELECT * FROM (VALUES
    (16::smallint,  0::smallint, 60691, 'ブラキディオス'),
    (16::smallint,  1::smallint, 60691, 'ブラキディオス'),
    (16::smallint,  2::smallint, 60691, 'ブラキディオス'),
    (16::smallint,  3::smallint, 60691, 'ブラキディオス'),
    (16::smallint,  4::smallint, 60691, 'ブラキディオス'),
    (16::smallint,  5::smallint, 60691, 'ブラキディオス'),
    (16::smallint,  6::smallint, 60691, 'ブラキディオス'),
    (16::smallint,  7::smallint, 60691, 'ブラキディオス'),
    (16::smallint,  8::smallint, 60691, 'ブラキディオス'),
    (16::smallint,  9::smallint, 60691, 'ブラキディオス'),
    (16::smallint, 10::smallint, 60691, 'ブラキディオス'),
    (16::smallint, 11::smallint, 60691, 'ブラキディオス'),
    (16::smallint, 12::smallint, 60691, 'ブラキディオス'),
    (16::smallint, 13::smallint, 60691, 'ブラキディオス'),
    (16::smallint, 14::smallint, 60691, 'ブラキディオス'),
    -- CupGroup 17: Guild hunt Brachydios (event_sub_type -1)
    (17::smallint, -1::smallint, 60690, 'ブラキディオスギルド'),
    -- CupGroup 6: Fishing size categories
    (6::smallint, 234::smallint, 0, 'キレアジ'),
    (6::smallint, 237::smallint, 0, 'ハリマグロ'),
    (6::smallint, 239::smallint, 0, 'カクサンデメキン')
) AS v(cup_group, event_sub_type, quest_file_id, name)
WHERE NOT EXISTS (SELECT 1 FROM tournament_sub_events);

-- Cups for the default tournament.
-- cup_type 7 = speed hunt, cup_type 6 = fishing size.
INSERT INTO tournament_cups (tournament_id, cup_group, cup_type, unk, name, description)
SELECT t.id, v.cup_group, v.cup_type, v.unk, v.name, v.description
FROM tournaments t
CROSS JOIN (VALUES
    (16::smallint, 7::smallint, 0::smallint, 'スピードハントカップ',   'ブラキディオスをより速く狩れ'),
    (17::smallint, 7::smallint, 0::smallint, 'ギルドハントカップ',     'ブラキディオスをギルドで狩れ'),
    (6::smallint,  6::smallint, 0::smallint, 'フィッシングサイズカップ', '大きな魚を釣れ')
) AS v(cup_group, cup_type, unk, name, description)
WHERE NOT EXISTS (SELECT 1 FROM tournament_cups WHERE tournament_id = t.id)
ORDER BY t.id;

COMMIT;
