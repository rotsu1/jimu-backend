package repository

const getRoutineSetByIDQuery = `
	SELECT id, routine_exercise_id, weight, reps, order_index, created_at, updated_at
	FROM public.routine_sets
	WHERE id = $1
`

const getRoutineSetsByRoutineExerciseIDQuery = `
	SELECT id, routine_exercise_id, weight, reps, order_index, created_at, updated_at
	FROM public.routine_sets
	WHERE routine_exercise_id = $1
	ORDER BY order_index ASC NULLS LAST, created_at ASC
`

const insertRoutineSetQuery = `
  INSERT INTO public.routine_sets (routine_exercise_id, weight, reps, order_index)
  SELECT $1, $2, $3, $4
  FROM public.routine_exercises re
  JOIN public.routines r ON re.routine_id = r.id
  WHERE re.id = $1 AND r.user_id = $5
  RETURNING id, routine_exercise_id, weight, reps, order_index, created_at, updated_at
`

const deleteRoutineSetByIDQuery = `
  DELETE FROM public.routine_sets
  WHERE id = $1
    AND routine_exercise_id IN (
        -- Guard: Set -> Exercise -> Routine -> User
        SELECT re.id 
        FROM public.routine_exercises re
        JOIN public.routines r ON re.routine_id = r.id
        WHERE (r.user_id = $2 OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2))
    )
`
