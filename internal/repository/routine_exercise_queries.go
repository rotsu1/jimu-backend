package repository

const getRoutineExerciseByIDQuery = `
	SELECT id, routine_id, exercise_id, order_index, rest_timer_seconds, memo, created_at, updated_at
	FROM public.routine_exercises
	WHERE id = $1
`

const getRoutineExercisesByRoutineIDQuery = `
	SELECT id, routine_id, exercise_id, order_index, rest_timer_seconds, memo, created_at, updated_at
	FROM public.routine_exercises
	WHERE routine_id = $1
	ORDER BY order_index ASC NULLS LAST, created_at ASC
`

const insertRoutineExerciseQuery = `
  INSERT INTO public.routine_exercises (routine_id, exercise_id, order_index, rest_timer_seconds, memo)
  SELECT $1, $2, $3, $4, $5
  FROM public.routines r
  WHERE r.id = $1 AND r.user_id = $6
  RETURNING id, routine_id, exercise_id, order_index, rest_timer_seconds, memo, created_at, updated_at
`

const deleteRoutineExerciseByIDQuery = `
  DELETE FROM public.routine_exercises
  WHERE id = $1
    AND routine_id IN (
        SELECT id FROM public.routines 
        WHERE (user_id = $2 OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2))
    )
`
