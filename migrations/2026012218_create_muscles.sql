-- +migrate Up
CREATE TABLE public.muscles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- +migrate Down
DROP TABLE public.muscles;