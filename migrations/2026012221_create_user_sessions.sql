-- +migrate Up
CREATE TABLE IF NOT EXISTS public.user_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES public.profiles(id) ON DELETE CASCADE,
    refresh_token text UNIQUE NOT NULL,
    user_agent text, -- Store device info (e.g., "iPhone 15 Pro")
    client_ip inet,
    is_revoked boolean NOT NULL DEFAULT false,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- Index for fast lookups during token refresh
CREATE INDEX idx_user_sessions_refresh_token ON public.user_sessions(refresh_token);

-- +migrate Down
DROP TABLE IF EXISTS public.user_sessions;