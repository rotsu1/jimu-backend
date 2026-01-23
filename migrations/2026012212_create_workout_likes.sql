-- +migrate Up
CREATE TABLE IF NOT EXISTS public.workout_likes (
    user_id uuid REFERENCES public.profiles(id) ON DELETE CASCADE,
    workout_id uuid REFERENCES public.workouts(id) ON DELETE CASCADE,
    created_at timestamp with time zone DEFAULT now(),
    PRIMARY KEY (user_id, workout_id)
);

CREATE INDEX IF NOT EXISTS idx_workout_likes_user_id ON public.workout_likes(user_id);
CREATE INDEX IF NOT EXISTS idx_workout_likes_workout_id ON public.workout_likes(workout_id);

-- +migrate Down
DROP TABLE IF EXISTS public.workout_likes;