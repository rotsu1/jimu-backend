--- +up
CREATE TABLE IF NOT EXISTS public.workout_exercises (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_id uuid REFERENCES public.workouts(id) ON DELETE CASCADE,
    exercise_id uuid REFERENCES public.exercises(id),
    order_index integer,
    memo text,
    rest_timer_seconds integer,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_workout_exercises_workout_id ON public.workout_exercises(workout_id);
CREATE INDEX IF NOT EXISTS idx_workout_exercises_exercise_id ON public.workout_exercises(exercise_id);

--- +down
DROP TABLE IF EXISTS public.workout_exercises;