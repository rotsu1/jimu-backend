-- +migrate Up
CREATE TABLE IF NOT EXISTS public.routine_sets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    routine_exercise_id uuid REFERENCES public.routine_exercises(id) ON DELETE CASCADE,
    weight numeric,
    reps integer,
    order_index integer NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_routine_sets_routine_exercise_id ON public.routine_sets(routine_exercise_id);

-- +migrate Down
DROP TABLE IF EXISTS public.routine_sets;