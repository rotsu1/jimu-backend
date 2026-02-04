-- +migrate Up
CREATE TABLE IF NOT EXISTS public.workout_sets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_exercise_id uuid REFERENCES public.workout_exercises(id) ON DELETE CASCADE,
    weight numeric,
    reps integer,
    order_index integer NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_workout_sets_workout_exercise_id ON public.workout_sets(workout_exercise_id);
CREATE INDEX IF NOT EXISTS idx_workout_exercises_workout_id ON public.workout_exercises(workout_id);

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

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

CREATE TRIGGER tr_sync_set_weight
AFTER INSERT OR UPDATE OR DELETE ON public.workout_sets
FOR EACH ROW EXECUTE FUNCTION handle_set_weight_sync();

-- +migrate Down
DROP TRIGGER IF EXISTS tr_sync_set_weight ON public.workout_sets;
DROP FUNCTION IF EXISTS handle_set_weight_sync;
DROP TABLE IF EXISTS public.workout_sets;