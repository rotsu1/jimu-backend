--- +up
INSERT INTO public.muscles (name)
VALUES 
    ('chest'), ('upper back'), ('lower back'), ('biceps'), 
    ('triceps'), ('shoulders'), ('forearm'), ('abs'), ('quads'), 
    ('hamstrings'), ('calf'), ('cardio'), ('other')
ON CONFLICT (name) DO NOTHING;

--- +down
DELETE FROM public.muscles WHERE name IN ('chest', 'upper back', 'lower back', 'biceps', 'triceps', 'shoulders', 'forearm', 'abs', 'quads', 'hamstrings', 'calf', 'cardio', 'other');