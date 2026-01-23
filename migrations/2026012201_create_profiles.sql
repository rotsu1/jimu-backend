--- +up
CREATE TABLE IF NOT EXISTS public.profiles (
    id uuid PRIMARY KEY,
    username text,
    primary_email text,
    display_name text,
    bio text,
    location text,
    birth_date date,
    avatar_url text,
    subscription_plan text, -- Adjusted from USER-DEFINED for compatibility
    is_private_account boolean DEFAULT false,
    last_worked_out_at timestamp with time zone,
    total_workouts integer DEFAULT 0,
    current_streak integer DEFAULT 0,
    total_weight integer DEFAULT 0,
    updated_at timestamp with time zone DEFAULT now(),
    created_at timestamp with time zone DEFAULT now()
);

-- Create index on username for faster lookups
CREATE INDEX IF NOT EXISTS idx_profiles_username ON public.profiles(username);

--- +down
DROP TABLE IF EXISTS public.profiles;