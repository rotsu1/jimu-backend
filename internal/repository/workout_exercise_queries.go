package repository

const getWorkoutExerciseByIDQuery = `
	SELECT id, workout_id, exercise_id, order_index, memo, rest_timer_seconds, created_at, updated_at
	FROM public.workout_exercises
	WHERE id = $1
`

const getWorkoutExercisesByWorkoutIDQuery = `
	SELECT id, workout_id, exercise_id, order_index, memo, rest_timer_seconds, created_at, updated_at
	FROM public.workout_exercises
	WHERE workout_id = $1
	ORDER BY order_index ASC NULLS LAST, created_at ASC
`

const insertWorkoutExerciseQuery = `
  INSERT INTO public.workout_exercises (workout_id, exercise_id, order_index, memo, rest_timer_seconds)
  SELECT $1, $2, $3, $4, $5
  WHERE EXISTS (
      -- Guard: Ensure the workout belongs to the user OR requester is an admin
      SELECT 1 FROM public.workouts 
      WHERE id = $1 
      AND (user_id = $6 OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $6))
  )
  RETURNING id, workout_id, exercise_id, order_index, memo, rest_timer_seconds, created_at, updated_at
`

const deleteWorkoutExerciseByIDQuery = `
  DELETE FROM public.workout_exercises
  WHERE id = $1
  AND workout_id IN (
      -- Guard: Only allow deletion if the workout belongs to the requester
      SELECT id FROM public.workouts 
      WHERE (user_id = $2 OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2))
  )
`
