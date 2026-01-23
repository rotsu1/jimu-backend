-- +migrate Up
CREATE TABLE IF NOT EXISTS public.routines (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid REFERENCES public.profiles(id) ON DELETE CASCADE,
    name text NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_routines_user_id ON public.routines(user_id);

-- +migrate Down
DROP TABLE IF EXISTS public.routines;