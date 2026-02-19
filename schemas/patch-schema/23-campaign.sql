BEGIN;

CREATE TABLE IF NOT EXISTS public.campaigns (
  id INTEGER PRIMARY KEY,
  min_hr INTEGER,
  max_hr INTEGER,
  min_sr INTEGER,
  max_sr INTEGER,
  min_gr INTEGER,
  max_gr INTEGER,
  reward_type INTEGER,
  stamps INTEGER,
  receive_type INTEGER,
  background_id INTEGER,
  start_time TIMESTAMP WITH TIME ZONE,
  end_time TIMESTAMP WITH TIME ZONE,
  title TEXT,
  reward TEXT,
  link TEXT,
  code_prefix TEXT
);

CREATE TABLE IF NOT EXISTS public.campaign_categories (
  id SERIAL PRIMARY KEY,
  type INTEGER,
  title TEXT,
  description TEXT
);

CREATE TABLE IF NOT EXISTS public.campaign_category_links (
  id SERIAL PRIMARY KEY,
  campaign_id INTEGER REFERENCES public.campaigns(id) ON DELETE CASCADE,
  category_id INTEGER REFERENCES public.campaign_categories(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.campaign_rewards (
  id SERIAL PRIMARY KEY,
  campaign_id INTEGER REFERENCES public.campaigns(id) ON DELETE CASCADE,
  item_type INTEGER,
  quantity INTEGER,
  item_id INTEGER
);

CREATE TABLE IF NOT EXISTS public.campaign_rewards_claimed (
  character_id INTEGER REFERENCES public.characters(id) ON DELETE CASCADE,
  reward_id INTEGER REFERENCES public.campaign_rewards(id) ON DELETE CASCADE,
  PRIMARY KEY (character_id, reward_id)
);

CREATE TABLE IF NOT EXISTS public.campaign_state (
  id SERIAL PRIMARY KEY,
  campaign_id INTEGER REFERENCES public.campaigns(id) ON DELETE CASCADE,
  character_id INTEGER REFERENCES public.characters(id) ON DELETE CASCADE,
  code TEXT
);

CREATE TABLE IF NOT EXISTS public.campaign_codes (
  code TEXT PRIMARY KEY,
  campaign_id INTEGER REFERENCES public.campaigns(id) ON DELETE CASCADE,
  multi BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS public.campaign_quest (
  campaign_id INTEGER REFERENCES public.campaigns(id) ON DELETE CASCADE,
  character_id INTEGER REFERENCES public.characters(id) ON DELETE CASCADE,
  PRIMARY KEY (campaign_id, character_id)
);

END;
