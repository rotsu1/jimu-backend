-- +migrate Up

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

-- +migrate StatementBegin
-- Create a function to sync the comments count on workouts
CREATE OR REPLACE FUNCTION handle_workout_comment_sync()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE public.workouts 
        SET comments_count = comments_count + 1 
        WHERE id = NEW.workout_id
        AND NEW.parent_id IS NULL;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE public.workouts 
        SET comments_count = comments_count - 1 
        WHERE id = OLD.workout_id
        AND OLD.parent_id IS NULL;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- Attach to comments
CREATE TRIGGER tr_sync_workout_comments
AFTER INSERT OR DELETE ON public.comments
FOR EACH ROW EXECUTE FUNCTION handle_workout_comment_sync();

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
AFTER INSERT OR DELETE ON public.workouts
FOR EACH ROW EXECUTE FUNCTION handle_profile_stats_sync();

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION handle_set_weight_sync()
RETURNS TRIGGER AS $$
DECLARE
    target_workout_id uuid;
    weight_diff numeric;
BEGIN
    -- 1. Find the workout_id associated with this set
    SELECT workout_id INTO target_workout_id 
    FROM public.workout_exercises 
    WHERE id = COALESCE(NEW.workout_exercise_id, OLD.workout_exercise_id);

    -- 2. Calculate the difference in weight
    -- We use COALESCE(..., 0) to handle NULLs safely
    IF (TG_OP = 'INSERT') THEN
        weight_diff := COALESCE(NEW.weight * NEW.reps, 0);
    ELSIF (TG_OP = 'DELETE') THEN
        weight_diff := -COALESCE(OLD.weight * OLD.reps, 0);
    ELSIF (TG_OP = 'UPDATE') THEN
        weight_diff := COALESCE(NEW.weight * NEW.reps, 0) - COALESCE(OLD.weight * OLD.reps, 0);
    END IF;

    -- 3. Update the workout total
    UPDATE public.workouts 
    SET total_weight = total_weight + weight_diff
    WHERE id = target_workout_id;

    -- 4. Update the profile total simultaneously!
    -- This keeps everything in sync in one single transaction
    UPDATE public.profiles
    SET total_weight = total_weight + weight_diff
    WHERE id = (SELECT user_id FROM public.workouts WHERE id = target_workout_id);

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

CREATE TRIGGER tr_sync_set_weight
AFTER INSERT OR UPDATE OR DELETE ON public.workout_sets
FOR EACH ROW EXECUTE FUNCTION handle_set_weight_sync();

-- +migrate Down
DROP FUNCTION IF EXISTS handle_workout_like_sync;
DROP FUNCTION IF EXISTS update_comment_like_count;
DROP FUNCTION IF EXISTS handle_workout_comment_sync;
DROP FUNCTION IF EXISTS handle_profile_stats_sync;
DROP FUNCTION IF EXISTS handle_set_weight_sync;
DROP TRIGGER IF EXISTS tr_sync_profile_stats ON public.workouts;
DROP TRIGGER IF EXISTS tr_sync_set_weight ON public.workout_sets;
DROP TRIGGER IF EXISTS tr_sync_workout_likes ON public.workout_likes;
DROP TRIGGER IF EXISTS tr_comment_likes_count ON public.comment_likes;
DROP TRIGGER IF EXISTS tr_sync_workout_comments ON public.comments;