-- +migrate Up
CREATE TABLE IF NOT EXISTS public.comment_likes (
    user_id uuid REFERENCES public.profiles(id) ON DELETE CASCADE,
    comment_id uuid REFERENCES public.comments(id) ON DELETE CASCADE,
    created_at timestamp with time zone DEFAULT now(),
    PRIMARY KEY (user_id, comment_id)
);

CREATE INDEX IF NOT EXISTS idx_comment_likes_user_id ON public.comment_likes(user_id);
CREATE INDEX IF NOT EXISTS idx_comment_likes_comment_id ON public.comment_likes(comment_id);

-- +migrate Down
DROP TABLE IF EXISTS public.comment_likes;