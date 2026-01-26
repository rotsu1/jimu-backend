package repository

const getWorkoutImageByIDQuery = `
	SELECT id, workout_id, storage_path, display_order, created_at, updated_at
	FROM public.workout_images
	WHERE id = $1
`

const getWorkoutImagesByWorkoutIDQuery = `
	SELECT id, workout_id, storage_path, display_order, created_at, updated_at
	FROM public.workout_images
	WHERE workout_id = $1
	ORDER BY display_order ASC NULLS LAST, created_at ASC
`

const insertWorkoutImageQuery = `
  INSERT INTO public.workout_images (workout_id, storage_path, display_order)
  SELECT $1, $2, $3
  WHERE EXISTS (
      -- Guard: Ensure the workout belongs to the user or is handled by an admin
      SELECT 1 FROM public.workouts 
      WHERE id = $1 
      AND (user_id = $4 OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $4))
  )
  RETURNING id, workout_id, storage_path, display_order, created_at, updated_at
`

const deleteWorkoutImageByIDQuery = `
  DELETE FROM public.workout_images
  WHERE id = $1
  AND workout_id IN (
      -- Guard: Verify ownership of the parent workout
      SELECT id FROM public.workouts 
      WHERE (user_id = $2 OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2))
  )
`
