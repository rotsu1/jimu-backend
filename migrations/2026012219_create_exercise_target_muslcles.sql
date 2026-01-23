--- +up
CREATE TABLE public.exercise_target_muscles (
    exercise_id uuid REFERENCES public.exercises(id) ON DELETE CASCADE,
    muscle_id uuid REFERENCES public.muscles(id) ON DELETE CASCADE,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    PRIMARY KEY (exercise_id, muscle_id)
);

--- +down
DROP TABLE public.exercise_target_muscles;