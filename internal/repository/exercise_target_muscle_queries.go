package repository

const getExerciseTargetMusclesByExerciseIDQuery = `
	SELECT exercise_id, muscle_id, created_at, updated_at
	FROM public.exercise_target_muscles
	WHERE exercise_id = $1
`

const insertExerciseTargetMuscleQuery = `
  INSERT INTO public.exercise_target_muscles (exercise_id, muscle_id)
  SELECT $1::uuid, $2::uuid
  WHERE EXISTS (
      -- Guard: The requester ($3) must own the exercise
      SELECT 1 FROM public.exercises 
      WHERE id = $1::uuid 
      AND user_id = $3::uuid 
  )
  ON CONFLICT (exercise_id, muscle_id) DO NOTHING
  RETURNING exercise_id, muscle_id, created_at, updated_at
`

const deleteExerciseTargetMusclesByExerciseIDQuery = `
  DELETE FROM public.exercise_target_muscles
  WHERE exercise_id = $1::uuid
    AND EXISTS (
        SELECT 1 FROM public.exercises 
        WHERE id = $1::uuid 
        AND (user_id = $2::uuid OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2::uuid))
    )
`

const deleteExerciseTargetMuscleQuery = `
  DELETE FROM public.exercise_target_muscles
  WHERE exercise_id = $1::uuid 
    AND muscle_id = $2::uuid
    AND EXISTS (
        SELECT 1 FROM public.exercises 
        WHERE id = $1::uuid 
        AND (user_id = $3::uuid OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $3::uuid))
    )
`
