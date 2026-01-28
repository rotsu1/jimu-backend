-- +migrate Up
CREATE TABLE public.exercise_target_muscles (
    exercise_id uuid REFERENCES public.exercises(id) ON DELETE CASCADE,
    muscle_id uuid REFERENCES public.muscles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (exercise_id, muscle_id)
);

-- +migrate Down
DROP TABLE public.exercise_target_muscles;