-- +migrate Up
CREATE TABLE public.blocked_users (
    blocker_id uuid REFERENCES public.profiles(id) ON DELETE CASCADE,
    blocked_id uuid REFERENCES public.profiles(id) ON DELETE CASCADE,
    created_at timestamp with time zone DEFAULT now(),
    PRIMARY KEY (blocker_id, blocked_id)
);

-- Index for checking "Is this user blocking me?" quickly
CREATE INDEX idx_blocked_lookup ON public.blocked_users(blocked_id, blocker_id);

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION public.fn_on_block_remove_follow()
RETURNS TRIGGER AS $$
BEGIN
    -- Remove follows in both directions
    DELETE FROM public.follows 
    WHERE (follower_id = NEW.blocker_id AND following_id = NEW.blocked_id)
       OR (follower_id = NEW.blocked_id AND following_id = NEW.blocker_id);
    RETURN NULL; 
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

CREATE TRIGGER tr_on_block_remove_follow
    AFTER INSERT ON public.blocked_users
    FOR EACH ROW
    EXECUTE FUNCTION public.fn_on_block_remove_follow();

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION public.fn_guard_follow_against_blocks()
RETURNS TRIGGER AS $$
BEGIN
    -- Check if User A has blocked User B OR if User B has blocked User A
    IF EXISTS (
        SELECT 1 FROM public.blocks 
        WHERE (blocker_id = NEW.follower_id AND blocked_id = NEW.following_id)
           OR (blocker_id = NEW.following_id AND blocked_id = NEW.follower_id)
    ) THEN
        -- This "RAISE EXCEPTION" cancels the entire transaction!
        RAISE EXCEPTION 'Cannot follow: a block exists between these users.';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

CREATE TRIGGER tr_guard_follow_against_blocks
    BEFORE INSERT OR UPDATE ON public.follows
    FOR EACH ROW
    EXECUTE FUNCTION public.fn_guard_follow_against_blocks();

-- +migrate Down
DROP TRIGGER IF EXISTS tr_on_block_remove_follow ON public.blocked_users
DROP TRIGGER IF EXISTS tr_guard_follow_against_blocks ON public.follows
DROP FUNCTION IF EXISTS public.fn_on_block_remove_follow
DROP FUNCTION IF EXISTS public.fn_guard_follow_against_blocks
DROP TABLE IF EXISTS public.blocked_users;