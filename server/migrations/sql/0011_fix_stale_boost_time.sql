-- Heal characters whose boost_time column holds a nonsensical value left
-- over from the pre-#187 bug. Two cases observed on live servers:
--
--   1. Pre-1970 timestamps (e.g. 1906-xx-xx) written when an uninitialised
--      time.Time wrapped through an int64->uint32 cast.
--   2. Far-future timestamps that are not consistent with any legitimate
--      BoostTimeDuration config (BoostTimeDuration is capped at a few hours).
--
-- Both are meaningless and should be NULL so that the boost system is
-- treated as inactive until the character triggers a fresh boost start.
-- NULL is the default for characters that have never boosted; see
-- handlers_cafe.go handleMsgMhfGetBoostTimeLimit / handleMsgMhfGetBoostRight.

UPDATE public.characters
   SET boost_time = NULL
 WHERE boost_time IS NOT NULL
   AND (boost_time < TIMESTAMP '1970-01-01 00:00:00'
        OR boost_time > now() + interval '10 years');
