-- +migrate Up
CREATE TABLE IF NOT EXISTS public.user_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES public.profiles(id) ON DELETE CASCADE,
    refresh_token text UNIQUE NOT NULL,
    user_agent text, -- Store device info (e.g., "iPhone 15 Pro")
    client_ip inet,
    is_revoked boolean DEFAULT false,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

-- Index for fast lookups during token refresh
CREATE INDEX idx_user_sessions_refresh_token ON public.user_sessions(refresh_token);

-- +migrate Down
DROP TABLE IF EXISTS public.user_sessions;