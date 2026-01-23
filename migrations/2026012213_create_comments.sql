-- +migrate Up
CREATE TABLE IF NOT EXISTS public.comments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid REFERENCES public.profiles(id) ON DELETE CASCADE,
    workout_id uuid REFERENCES public.workouts(id) ON DELETE CASCADE,
    parent_id uuid REFERENCES public.comments(id) ON DELETE CASCADE,
    content text NOT NULL,
    likes_count integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now()
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON public.comments(user_id);
CREATE INDEX IF NOT EXISTS idx_comments_workout_id ON public.comments(workout_id);
CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON public.comments(parent_id);
CREATE INDEX IF NOT EXISTS idx_comments_created_at ON public.comments(created_at);
CREATE INDEX IF NOT EXISTS idx_workouts_user_id ON public.workouts(user_id);
CREATE INDEX IF NOT EXISTS idx_profiles_privacy_status ON public.profiles(id, is_private_account);

-- +migrate Down
DROP TABLE IF EXISTS public.comments;