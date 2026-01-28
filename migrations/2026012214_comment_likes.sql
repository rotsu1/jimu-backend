-- +migrate Up
CREATE TABLE IF NOT EXISTS public.comment_likes (
    user_id uuid REFERENCES public.profiles(id) ON DELETE CASCADE,
    comment_id uuid REFERENCES public.comments(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (user_id, comment_id)
);

CREATE INDEX IF NOT EXISTS idx_comment_likes_user_id ON public.comment_likes(user_id);
CREATE INDEX IF NOT EXISTS idx_comment_likes_comment_id ON public.comment_likes(comment_id);

-- +migrate StatementBegin
-- Create a function to update the comment like count
CREATE OR REPLACE FUNCTION update_comment_like_count()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE public.comments SET likes_count = likes_count + 1 WHERE id = NEW.comment_id;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE public.comments SET likes_count = likes_count - 1 WHERE id = OLD.comment_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- Trigger for Comment Likes
CREATE TRIGGER tr_comment_likes_count
AFTER INSERT OR DELETE ON public.comment_likes
FOR EACH ROW EXECUTE FUNCTION update_comment_like_count();

-- +migrate Down
DROP TRIGGER IF EXISTS tr_comment_likes_count ON public.comment_likes;
DROP FUNCTION IF EXISTS update_comment_like_count;
DROP TABLE IF EXISTS public.comment_likes;