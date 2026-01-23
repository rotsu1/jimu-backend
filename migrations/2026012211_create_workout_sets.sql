--- +up
CREATE TABLE IF NOT EXISTS public.workout_sets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_exercise_id uuid REFERENCES public.workout_exercises(id) ON DELETE CASCADE,
    weight numeric,
    reps integer,
    is_completed boolean DEFAULT false,
    order_index integer,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_workout_sets_workout_exercise_id ON public.workout_sets(workout_exercise_id);
CREATE INDEX IF NOT EXISTS idx_workout_exercises_workout_id ON public.workout_exercises(workout_id);

--- +down
DROP TABLE IF EXISTS public.workout_sets;