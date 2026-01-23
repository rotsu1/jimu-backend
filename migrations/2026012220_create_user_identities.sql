--- +up
CREATE TABLE IF NOT EXISTS public.user_identities (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES public.profiles(id) ON DELETE CASCADE,
    provider_name text NOT NULL, -- 'google' or 'apple'
    provider_user_id text NOT NULL, -- The 'sub' or 'id' from the Google/Apple token
    provider_email text,
    last_sign_in_at timestamp with time zone DEFAULT now(),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    UNIQUE(provider_name, provider_user_id)
);

--- +down
DROP TABLE IF EXISTS public.user_identities;