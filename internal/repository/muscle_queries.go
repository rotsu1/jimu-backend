package repository

const getMuscleByIDQuery = `
	SELECT id, name, created_at
	FROM public.muscles
	WHERE id = $1
`

const getAllMusclesQuery = `
	SELECT id, name, created_at
	FROM public.muscles
	ORDER BY name ASC
`

const getMuscleByNameQuery = `
	SELECT id, name, created_at
	FROM public.muscles
	WHERE name = $1
`
const createMuscleQuery = `
  INSERT INTO public.muscles (name)
  SELECT $1
  WHERE EXISTS (
      -- Guard: Only allow the insert if the requester is an admin
      SELECT 1 FROM public.sys_admins WHERE user_id = $2
  )
  RETURNING id, name, created_at
`

const deleteMuscleQuery = `
  DELETE FROM public.muscles
  WHERE id = $1
  AND EXISTS (
      -- Guard: Only allow deletion if the requester is in the sys_admins table
      SELECT 1 FROM public.sys_admins WHERE user_id = $2
  )
`
