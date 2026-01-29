-- +migrate Up
CREATE TABLE IF NOT EXISTS public.workout_images (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_id uuid REFERENCES public.workouts(id) ON DELETE CASCADE,
    storage_path text,
    display_order integer NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_workout_images_workout_id ON public.workout_images(workout_id);

-- +migrate Down
DROP TABLE IF EXISTS public.workout_images;