-- +migrate Up
CREATE TABLE IF NOT EXISTS public.follows (
    follower_id uuid REFERENCES public.profiles(id) ON DELETE CASCADE,
    following_id uuid REFERENCES public.profiles(id) ON DELETE CASCADE,
    status text DEFAULT 'pending', -- 'pending', 'accepted'
    created_at timestamp with time zone DEFAULT now(),
    PRIMARY KEY (follower_id, following_id)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_follows_follower_id ON public.follows(follower_id);
CREATE INDEX IF NOT EXISTS idx_follows_following_id ON public.follows(following_id);

-- +migrate Down
DROP TABLE IF EXISTS public.follows;