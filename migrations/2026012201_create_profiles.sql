-- +migrate Up
CREATE TABLE IF NOT EXISTS public.profiles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username text UNIQUE,
    display_name text,
    bio text,
    location text,
    birth_date date,
    avatar_url text,
    subscription_plan text,
    is_private_account boolean DEFAULT false,
    last_worked_out_at timestamp with time zone,
    total_workouts integer DEFAULT 0,
    current_streak integer DEFAULT 0,
    total_weight integer DEFAULT 0,
    followers_count integer DEFAULT 0,
    following_count integer DEFAULT 0,
    updated_at timestamp with time zone DEFAULT now(),
    created_at timestamp with time zone DEFAULT now()
);

-- Create index on username for faster lookups
CREATE INDEX IF NOT EXISTS idx_profiles_username ON public.profiles(username);

-- +migrate Down
DROP TABLE IF EXISTS public.profiles;