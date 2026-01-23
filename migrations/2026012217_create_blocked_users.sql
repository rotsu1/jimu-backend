--- +up
CREATE TABLE public.blocked_users (
    blocker_id uuid REFERENCES public.profiles(id) ON DELETE CASCADE,
    blocked_id uuid REFERENCES public.profiles(id) ON DELETE CASCADE,
    created_at timestamp with time zone DEFAULT now(),
    PRIMARY KEY (blocker_id, blocked_id)
);

-- Index for checking "Is this user blocking me?" quickly
CREATE INDEX idx_blocked_lookup ON public.blocked_users(blocked_id, blocker_id);

--- +down
DROP TABLE IF EXISTS public.blocked_users;