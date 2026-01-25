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

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION public.fn_on_follow_sync()
RETURNS TRIGGER AS $$
BEGIN
    -- 1. Handle NEW FOLLOWS (Only if they are auto-accepted or public)
    IF (TG_OP = 'INSERT') THEN
        IF (NEW.status = 'accepted') THEN
            UPDATE public.profiles SET followers_count = followers_count + 1 WHERE id = NEW.following_id;
            UPDATE public.profiles SET following_count = following_count + 1 WHERE id = NEW.follower_id;
        END IF;

    -- 2. Handle ACCEPTING a pending request (The status changes)
    ELSIF (TG_OP = 'UPDATE') THEN
        IF (OLD.status = 'pending' AND NEW.status = 'accepted') THEN
            UPDATE public.profiles SET followers_count = followers_count + 1 WHERE id = NEW.following_id;
            UPDATE public.profiles SET following_count = following_count + 1 WHERE id = NEW.follower_id;
        END IF;

    -- 3. Handle UNFOLLOWING or REJECTING
    ELSIF (TG_OP = 'DELETE') THEN
        -- Only decrement if the follow was actually accepted/active
        IF (OLD.status = 'accepted') THEN
            UPDATE public.profiles SET followers_count = followers_count - 1 WHERE id = OLD.following_id;
            UPDATE public.profiles SET following_count = following_count - 1 WHERE id = OLD.follower_id;
        END IF;
    END IF;
    
    RETURN NULL; -- AFTER triggers can return NULL
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- Attach to all three actions!
CREATE TRIGGER tr_sync_follow_counts
    AFTER INSERT OR UPDATE OR DELETE ON public.follows
    FOR EACH ROW
    EXECUTE FUNCTION public.fn_on_follow_sync();

-- +migrate Down
DROP TRIGGER IF EXISTS tr_sync_follow_counts ON public.follows;
DROP FUNCTION IF EXISTS public.fn_on_follow_sync;
DROP TABLE IF EXISTS public.follows;