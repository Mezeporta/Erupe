DO $$ BEGIN
    -- Only apply if the new-schema columns exist (item_type vs legacy itemtype)
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name='fpoint_items' AND column_name='item_type'
    ) THEN
        DELETE FROM public.fpoint_items;
        ALTER TABLE public.fpoint_items ALTER COLUMN item_type SET NOT NULL;
        ALTER TABLE public.fpoint_items ALTER COLUMN item_id SET NOT NULL;
        ALTER TABLE public.fpoint_items ALTER COLUMN quantity SET NOT NULL;
        ALTER TABLE public.fpoint_items ALTER COLUMN fpoints SET NOT NULL;
        ALTER TABLE public.fpoint_items DROP COLUMN IF EXISTS trade_type;
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_name='fpoint_items' AND column_name='buyable'
        ) THEN
            ALTER TABLE public.fpoint_items ADD COLUMN buyable boolean NOT NULL DEFAULT false;
        END IF;
    END IF;
END $$;