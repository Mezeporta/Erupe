-- Add unique constraint required for ON CONFLICT upsert in RecordPurchase.
-- Uses CREATE UNIQUE INDEX which supports IF NOT EXISTS, avoiding errors
-- when the baseline schema (0001) already includes the constraint.
CREATE UNIQUE INDEX IF NOT EXISTS shop_items_bought_character_item_unique
    ON public.shop_items_bought (character_id, shop_item_id);
