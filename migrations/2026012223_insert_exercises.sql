-- +migrate Up

-- 1. Insert ALL Exercises at once (Much faster and cleaner)
INSERT INTO public.exercises (name, user_id, icon)
VALUES 
    -- Chest
    ('Bench Press', NULL, 'chest'),
    ('Incline Bench Press', NULL, 'chest'),
    ('Dumbbell Bench Press', NULL, 'chest'),
    ('Incline Dumbbell Bench Press', NULL, 'chest'),
    ('Dumbbell Fly', NULL, 'chest'),
    ('Incline Dumbbell Fly', NULL, 'chest'),
    ('Cable Fly', NULL, 'chest'),
    -- Back
    ('Pull Up', NULL, 'back'),
    ('Bent Over Row', NULL, 'back'),
    ('Lat Pulldown', NULL, 'back'),
    ('Deadlift', NULL, 'back'),
    ('Back Extension', NULL, 'back'),
    -- Legs
    ('Squat', NULL, 'legs'),
    ('Leg Press', NULL, 'legs'),
    ('Leg Extension', NULL, 'legs'),
    ('Leg Curl', NULL, 'legs'),
    ('Standing Calf Raise', NULL, 'legs'),
    -- Arms
    ('Barbell Curl', NULL, 'arms'),
    ('Hammer Curl', NULL, 'arms'),
    ('Dumbell Curl', NULL, 'arms'),
    ('Tricep Pushdown', NULL, 'arms'),
    ('Skull Crusher', NULL, 'arms'),
    ('Wrist Curl', NULL, 'arms'),
    ('Narrow Grip Bench Press', NULL, 'arms'),
    -- Core & Cardio
    ('Plank', NULL, 'core'),
    ('Crunches', NULL, 'core'),
    ('Running', NULL, 'heart'),
    ('Cycling', NULL, 'heart'),
    -- Shoulders
    ('Dumbell Shoulder Press', NULL, 'shoulders'),
    ('Barbell Shoulder Press', NULL, 'shoulders'),
    ('Lateral Raise', NULL, 'shoulders'),
    ('Front Raise', NULL, 'shoulders'),
    ('Rear Fly', NULL, 'shoulders'),
    ('Shoulder Shrug', NULL, 'shoulders'),
    -- Other
    ('Stretching', NULL, 'other'),
    ('Yoga', NULL, 'other'),
    ('Pilates', NULL, 'other'),
    ('Meditation', NULL, 'other'),
    ('Breathing Exercises', NULL, 'other');

-- 2. Link CHEST Exercises
INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
SELECT e.id, m.id 
FROM public.exercises e, public.muscles m 
WHERE m.name = 'chest' 
AND e.name IN ('Bench Press', 'Incline Bench Press', 'Dumbbell Bench Press', 'Incline Dumbbell Bench Press', 'Dumbbell Fly', 'Incline Dumbbell Fly', 'Cable Fly');

-- 3. Link BACK Exercises (Upper Back for all, Lower Back for specifically Deadlift/Extension)
INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
SELECT e.id, m.id 
FROM public.exercises e, public.muscles m 
WHERE m.name = 'upper back' 
AND e.name IN ('Pull Up', 'Bent Over Row', 'Lat Pulldown', 'Deadlift', 'Back Extension');

INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
SELECT e.id, m.id 
FROM public.exercises e, public.muscles m 
WHERE m.name = 'lower back' 
AND e.name IN ('Deadlift', 'Back Extension');

-- 4. Link LEG Exercises
-- Quads
INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
SELECT e.id, m.id FROM public.exercises e, public.muscles m 
WHERE m.name = 'quads' AND e.name IN ('Squat', 'Leg Press', 'Leg Extension');
-- Hamstrings
INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
SELECT e.id, m.id FROM public.exercises e, public.muscles m 
WHERE m.name = 'hamstrings' AND e.name = 'Leg Curl';
-- Calves
INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
SELECT e.id, m.id FROM public.exercises e, public.muscles m 
WHERE m.name = 'calf' AND e.name = 'Standing Calf Raise';

-- 5. Link ARM Exercises
-- Biceps (Any "Curl" except Wrist Curl)
INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
SELECT e.id, m.id FROM public.exercises e, public.muscles m 
WHERE m.name = 'biceps' AND e.name LIKE '%Curl%' AND e.name != 'Wrist Curl';
-- Triceps
INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
SELECT e.id, m.id FROM public.exercises e, public.muscles m 
WHERE m.name = 'triceps' AND e.name IN ('Tricep Pushdown', 'Skull Crusher', 'Narrow Grip Bench Press');
-- Forearms
INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
SELECT e.id, m.id FROM public.exercises e, public.muscles m 
WHERE m.name = 'forearm' AND e.name = 'Wrist Curl';

-- 6. Link CORE & CARDIO
INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
SELECT e.id, m.id FROM public.exercises e, public.muscles m 
WHERE m.name = 'abs' AND e.name IN ('Plank', 'Crunches');

INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
SELECT e.id, m.id FROM public.exercises e, public.muscles m 
WHERE m.name = 'cardio' AND e.name IN ('Running', 'Cycling');

-- 7. Link SHOULDERS (All shoulder exercises)
INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
SELECT e.id, m.id FROM public.exercises e, public.muscles m 
WHERE m.name = 'shoulders' 
AND e.name IN ('Dumbell Shoulder Press', 'Barbell Shoulder Press', 'Lateral Raise', 'Front Raise', 'Rear Fly', 'Shoulder Shrug');

-- 8. Link OTHER
INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
SELECT e.id, m.id FROM public.exercises e, public.muscles m 
WHERE m.name = 'other' 
AND e.name IN ('Stretching', 'Yoga', 'Pilates', 'Meditation', 'Breathing Exercises');

-- +migrate Down
-- Delete the links first (Foreign Key safety)
DELETE FROM public.exercise_target_muscles 
WHERE exercise_id IN (
    SELECT id FROM public.exercises WHERE name IN (
        'Bench Press', 'Incline Bench Press', 'Dumbbell Bench Press', 'Incline Dumbbell Bench Press', 
        'Dumbbell Fly', 'Incline Dumbbell Fly', 'Cable Fly', 'Pull Up', 'Bent Over Row', 'Lat Pulldown', 
        'Deadlift', 'Back Extension', 'Squat', 'Leg Press', 'Leg Extension', 'Leg Curl', 'Standing Calf Raise', 
        'Barbell Curl', 'Hammer Curl', 'Dumbell Curl', 'Tricep Pushdown', 'Skull Crusher', 'Wrist Curl', 
        'Narrow Grip Bench Press', 'Plank', 'Crunches', 'Running', 'Cycling', 'Dumbell Shoulder Press', 
        'Barbell Shoulder Press', 'Lateral Raise', 'Front Raise', 'Rear Fly', 'Shoulder Shrug', 'Stretching', 
        'Yoga', 'Pilates', 'Meditation', 'Breathing Exercises'
    )
);

-- Then delete the exercises
DELETE FROM public.exercises WHERE name IN (
  'Bench Press', 'Incline Bench Press', 'Dumbbell Bench Press', 'Incline Dumbbell Bench Press', 
  'Dumbbell Fly', 'Incline Dumbbell Fly', 'Cable Fly', 'Pull Up', 'Bent Over Row', 'Lat Pulldown', 
  'Deadlift', 'Back Extension', 'Squat', 'Leg Press', 'Leg Extension', 'Leg Curl', 'Standing Calf Raise', 
  'Barbell Curl', 'Hammer Curl', 'Dumbell Curl', 'Tricep Pushdown', 'Skull Crusher', 'Wrist Curl', 
  'Narrow Grip Bench Press', 'Plank', 'Crunches', 'Running', 'Cycling', 'Dumbell Shoulder Press', 
  'Barbell Shoulder Press', 'Lateral Raise', 'Front Raise', 'Rear Fly', 'Shoulder Shrug', 'Stretching', 
  'Yoga', 'Pilates', 'Meditation', 'Breathing Exercises'
);