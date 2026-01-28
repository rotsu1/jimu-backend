-- +migrate Up
CREATE TABLE IF NOT EXISTS public.user_settings (
    user_id uuid PRIMARY KEY REFERENCES public.profiles(id) ON DELETE CASCADE,
    notify_new_follower boolean DEFAULT true,
    notify_likes boolean DEFAULT true,
    notify_comments boolean DEFAULT true,
    sound_enabled boolean DEFAULT true,
    sound_effect_name text DEFAULT 'bell', 
    default_timer_seconds integer DEFAULT 120,
    auto_fill_previous_values boolean DEFAULT true,
    unit_weight text DEFAULT 'kg',
    unit_distance text DEFAULT 'km',
    unit_length text DEFAULT 'cm',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- +migrate Down
DROP TABLE IF EXISTS public.user_settings;