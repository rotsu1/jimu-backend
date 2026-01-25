-- +migrate Up
CREATE TABLE IF NOT EXISTS public.workout_likes (
    user_id uuid REFERENCES public.profiles(id) ON DELETE CASCADE,
    workout_id uuid REFERENCES public.workouts(id) ON DELETE CASCADE,
    created_at timestamp with time zone DEFAULT now(),
    PRIMARY KEY (user_id, workout_id)
);

CREATE INDEX IF NOT EXISTS idx_workout_likes_user_id ON public.workout_likes(user_id);
CREATE INDEX IF NOT EXISTS idx_workout_likes_workout_id ON public.workout_likes(workout_id);

-- +migrate StatementBegin
-- Create a function to sync the likes count on workout_likes
CREATE OR REPLACE FUNCTION handle_workout_like_sync()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE public.workouts 
        SET likes_count = likes_count + 1 
        WHERE id = NEW.workout_id;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE public.workouts 
        SET likes_count = likes_count - 1 
        WHERE id = OLD.workout_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- Attach to workout_likes
CREATE TRIGGER tr_sync_workout_likes
AFTER INSERT OR DELETE ON public.workout_likes
FOR EACH ROW EXECUTE FUNCTION handle_workout_like_sync();

-- +migrate Down
DROP TRIGGER IF EXISTS tr_sync_workout_likes ON public.workout_likes;
DROP FUNCTION IF EXISTS handle_workout_like_sync;
DROP TABLE IF EXISTS public.workout_likes;