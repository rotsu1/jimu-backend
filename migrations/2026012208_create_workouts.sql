-- +migrate Up
CREATE TABLE IF NOT EXISTS public.workouts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid REFERENCES public.profiles(id) ON DELETE CASCADE,
    name text,
    comment text,
    started_at timestamp with time zone NOT NULL,
    ended_at timestamp with time zone NOT NULL,
    duration_seconds integer DEFAULT 0 NOT NULL,
    total_weight integer DEFAULT 0 NOT NULL,
    likes_count integer DEFAULT 0 NOT NULL,
    comments_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_workouts_user_id ON public.workouts(user_id);
CREATE INDEX IF NOT EXISTS idx_workouts_started_at ON public.workouts(started_at);
CREATE INDEX IF NOT EXISTS idx_workouts_created_at ON public.workouts(created_at);

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION handle_profile_stats_sync()
RETURNS TRIGGER AS $$
DECLARE
    last_workout timestamp with time zone;
BEGIN
    IF (TG_OP = 'INSERT') THEN
        -- 1. Get the last workout time before we update it
        SELECT last_worked_out_at INTO last_workout 
        FROM public.profiles 
        WHERE id = NEW.user_id;

        UPDATE public.profiles
        SET 
            total_workouts = total_workouts + 1,
            total_weight = total_weight + COALESCE(NEW.total_weight, 0),
            
            -- 2. Logic for current_streak
            current_streak = CASE
                -- Case: First workout ever
                WHEN last_workout IS NULL THEN 1
                
                -- Case: Worked out yesterday (Increment!)
                -- We check if the difference is exactly 1 day (ignoring time)
                WHEN (NEW.started_at::date - last_workout::date) = 1 THEN current_streak + 1
                
                -- Case: Already worked out today (Keep same)
                WHEN (NEW.started_at::date - last_workout::date) = 0 THEN current_streak
                
                -- Case: Missed a day (Reset)
                ELSE 1
            END,
            
            last_worked_out_at = NEW.started_at
        WHERE id = NEW.user_id;

    ELSIF (TG_OP = 'UPDATE') THEN
        -- Only update profile if the weight actually changed
        IF NEW.total_weight <> OLD.total_weight THEN
            UPDATE public.profiles
            SET total_weight = total_weight - OLD.total_weight + NEW.total_weight
            WHERE id = NEW.user_id;
        END IF;

    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE public.profiles
        SET 
            total_workouts = total_workouts - 1,
            total_weight = total_weight - COALESCE(OLD.total_weight, 0)
            -- We typically don't recalculate streaks on a delete 
            -- because it's computationally expensive!
        WHERE id = OLD.user_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- Attach to workouts
CREATE TRIGGER tr_sync_profile_stats
AFTER INSERT OR UPDATE OR DELETE ON public.workouts
FOR EACH ROW EXECUTE FUNCTION handle_profile_stats_sync();

-- +migrate Down
DROP TRIGGER IF EXISTS tr_sync_profile_stats ON public.workouts;
DROP FUNCTION IF EXISTS handle_profile_stats_sync;
DROP TABLE IF EXISTS public.workouts;