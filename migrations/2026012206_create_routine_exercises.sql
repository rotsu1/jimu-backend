-- +migrate Up
CREATE TABLE IF NOT EXISTS public.routine_exercises (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    routine_id uuid REFERENCES public.routines(id) ON DELETE CASCADE,
    exercise_id uuid REFERENCES public.exercises(id),
    order_index integer,
    rest_timer_seconds integer,
    memo text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_routine_exercises_routine_id ON public.routine_exercises(routine_id);
CREATE INDEX IF NOT EXISTS idx_routine_exercises_exercise_id ON public.routine_exercises(exercise_id);

-- +migrate Down
DROP TABLE IF EXISTS public.routine_exercises;