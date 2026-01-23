--- +up
WITH inserted_exercises AS (
    -- Insert the exercises and return their IDs and names
    INSERT INTO public.exercises (name, user_id, icon)
    VALUES 
        ('Bench Press', NULL, 'chest'),
        ('Incline Bench Press', NULL, 'chest'),
        ('Dumbbell Bench Press', NULL, 'chest'),
        ('Incline Dumbbell Bench Press', NULL, 'chest'),
        ('Dumbbell Fly', NULL, 'chest'),
        ('Incline Dumbbell Fly', NULL, 'chest'),
        ('Cable Fly', NULL, 'chest')
    RETURNING id, name
)
-- Link all these new exercises to the 'chest' muscle
INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
SELECT 
    ie.id, 
    (SELECT id FROM public.muscles WHERE name = 'chest')
FROM inserted_exercises ie;

-- 1. BACK (Upper and Lower)
WITH back_exercises AS (
    INSERT INTO public.exercises (name, user_id, icon)
    VALUES 
        ('Pull Up', NULL, 'back'),
        ('Bent Over Row', NULL, 'back'),
        ('Lat Pulldown', NULL, 'back'),
        ('Deadlift', NULL, 'back'),
        ('Back Extension', NULL, 'back')
    RETURNING id, name
)
INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
SELECT id, (SELECT id FROM public.muscles WHERE name = 'upper back') FROM back_exercises
UNION ALL
SELECT id, (SELECT id FROM public.muscles WHERE name = 'lower back') FROM back_exercises WHERE name IN ('Deadlift', 'Back Extension');

-- 2. LEGS (Quads, Hamstrings, Calves)
WITH leg_exercises AS (
    INSERT INTO public.exercises (name, user_id, icon)
    VALUES 
        ('Squat', NULL, 'legs'),
        ('Leg Press', NULL, 'legs'),
        ('Leg Extension', NULL, 'legs'),
        ('Leg Curl', NULL, 'legs'),
        ('Standing Calf Raise', NULL, 'legs')
    RETURNING id, name
)
INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
SELECT id, (SELECT id FROM public.muscles WHERE name = 'quads') FROM leg_exercises WHERE name IN ('Squat', 'Leg Press', 'Leg Extension')
UNION ALL
SELECT id, (SELECT id FROM public.muscles WHERE name = 'hamstrings') FROM leg_exercises WHERE name = 'Leg Curl'
UNION ALL
SELECT id, (SELECT id FROM public.muscles WHERE name = 'calf') FROM leg_exercises WHERE name = 'Standing Calf Raise';

-- 3. ARMS (Biceps, Triceps, Forearms)
WITH arm_exercises AS (
    INSERT INTO public.exercises (name, user_id, icon)
    VALUES 
        ('Barbell Curl', NULL, 'arms'),
        ('Hammer Curl', NULL, 'arms'),
        ('Dumbell Curl', NULL, 'arms'),
        ('Tricep Pushdown', NULL, 'arms'),
        ('Skull Crusher', NULL, 'arms'),
        ('Wrist Curl', NULL, 'arms'),
        ('Narrow Grip Bench Press', NULL, 'arms'),
    RETURNING id, name
)
INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
SELECT id, (SELECT id FROM public.muscles WHERE name = 'biceps') FROM arm_exercises WHERE name LIKE '%Curl%' AND name != 'Wrist Curl'
UNION ALL
SELECT id, (SELECT id FROM public.muscles WHERE name = 'triceps') FROM arm_exercises WHERE name IN ('Tricep Pushdown', 'Skull Crusher', 'Narrow Grip Bench Press')
UNION ALL
SELECT id, (SELECT id FROM public.muscles WHERE name = 'forearm') FROM arm_exercises WHERE name = 'Wrist Curl';

-- 4. CORE & CARDIO
WITH misc_exercises AS (
    INSERT INTO public.exercises (name, user_id, icon)
    VALUES 
        ('Plank', NULL, 'core'),
        ('Crunches', NULL, 'core'),
        ('Running', NULL, 'heart'),
        ('Cycling', NULL, 'heart')
    RETURNING id, name
)
INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
SELECT id, (SELECT id FROM public.muscles WHERE name = 'abs') FROM misc_exercises WHERE name IN ('Plank', 'Crunches')
UNION ALL
SELECT id, (SELECT id FROM public.muscles WHERE name = 'cardio') FROM misc_exercises WHERE name IN ('Running', 'Cycling');

-- 5. SHOULDERS
WITH shoulder_exercises AS (
    INSERT INTO public.exercises (name, user_id, icon)
    VALUES 
        ('Dumbell Shoulder Press', NULL, 'shoulders'),
        ('Barbell Shoulder Press', NULL, 'shoulders'),
        ('Lateral Raise', NULL, 'shoulders'),
        ('Front Raise', NULL, 'shoulders'),
        ('Rear Fly', NULL, 'shoulders'),
        ('Shoulder Shrug', NULL, 'shoulders')
    RETURNING id, name
)
INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
SELECT id, (SELECT id FROM public.muscles WHERE name = 'shoulders') FROM shoulder_exercises;

-- 7. OTHER
WITH other_exercises AS (
    INSERT INTO public.exercises (name, user_id, icon)
    VALUES 
        ('Stretching', NULL, 'other'),
        ('Yoga', NULL, 'other'),
        ('Pilates', NULL, 'other'),
        ('Meditation', NULL, 'other'),
        ('Breathing Exercises', NULL, 'other'),
    RETURNING id, name
)
INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
SELECT id, (SELECT id FROM public.muscles WHERE name = 'other') FROM other_exercises;

--- +down
DELETE FROM public.exercise_target_muscles WHERE exercise_id IN (SELECT id FROM public.exercises);
DELETE FROM public.exercises WHERE name IN (
  'Bench Press', 'Incline Bench Press', 'Dumbbell Bench Press', 'Incline Dumbbell Bench Press', 
  'Dumbbell Fly', 'Incline Dumbbell Fly', 'Cable Fly', 'Pull Up', 'Bent Over Row', 'Lat Pulldown', 
  'Deadlift', 'Back Extension', 'Squat', 'Leg Press', 'Leg Extension', 'Leg Curl', 'Standing Calf Raise', 
  'Barbell Curl', 'Hammer Curl', 'Dumbell Curl', 'Tricep Pushdown', 'Skull Crusher', 'Wrist Curl', 
  'Narrow Grip Bench Press', 'Plank', 'Crunches', 'Running', 'Cycling', 'Dumbell Shoulder Press', 
  'Barbell Shoulder Press', 'Lateral Raise', 'Front Raise', 'Rear Fly', 'Shoulder Shrug', 'Stretching', 
  'Yoga', 'Pilates', 'Meditation', 'Breathing Exercises'
);

