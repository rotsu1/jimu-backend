--- +up
CREATE TABLE public.exercise_target_muscles (
    exercise_id uuid REFERENCES public.exercises(id) ON DELETE CASCADE,
    muscle_id uuid REFERENCES public.muscles(id) ON DELETE CASCADE,
    PRIMARY KEY (exercise_id, muscle_id)
);

--- +down
DROP TABLE public.exercise_target_muscles;