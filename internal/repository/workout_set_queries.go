package repository

const getWorkoutSetByIDQuery = `
	SELECT id, workout_exercise_id, weight, reps, is_completed, order_index, created_at, updated_at
	FROM public.workout_sets
	WHERE id = $1
`

const getWorkoutSetsByWorkoutExerciseIDQuery = `
	SELECT id, workout_exercise_id, weight, reps, is_completed, order_index, created_at, updated_at
	FROM public.workout_sets
	WHERE workout_exercise_id = $1
	ORDER BY order_index ASC NULLS LAST, created_at ASC
`

const insertWorkoutSetQuery = `
  INSERT INTO public.workout_sets (workout_exercise_id, weight, reps, is_completed, order_index)
  SELECT $1, $2, $3, $4, $5
  WHERE EXISTS (
      -- Guard: Ensure the parent exercise belongs to a workout owned by this user
      SELECT 1 FROM public.workout_exercises we
      JOIN public.workouts w ON we.workout_id = w.id
      WHERE we.id = $1 
      AND (w.user_id = $6 OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $6))
  )
  RETURNING id, workout_exercise_id, weight, reps, is_completed, order_index, created_at, updated_at
`

const deleteWorkoutSetByIDQuery = `
  DELETE FROM public.workout_sets
  WHERE id = $1
  AND workout_exercise_id IN (
      -- Guard: Only allow deletion if the hierarchy leads back to the owner
      SELECT we.id FROM public.workout_exercises we
      JOIN public.workouts w ON we.workout_id = w.id
      WHERE (w.user_id = $2 OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2))
  )
`
