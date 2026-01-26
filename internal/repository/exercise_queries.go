package repository

const getExerciseByIDQuery = `
  SELECT id, user_id, name, suggested_rest_seconds, icon, created_at, updated_at
  FROM public.exercises e
  WHERE e.id = $1
    AND (
      -- 1. System exercises (NULL user_id) are always public
      e.user_id IS NULL 
      -- 2. Owner or Admin always has access
      OR e.user_id = $2 
      OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2)
      -- 3. Social Visibility via Workout
      OR EXISTS (
          SELECT 1 FROM public.workout_exercises we
          JOIN public.workouts w ON we.workout_id = w.id
          JOIN public.profiles p ON w.user_id = p.id
          WHERE we.exercise_id = e.id 
            AND w.user_id != $2 -- Don't check blocks against yourself
            -- GUARD: Requester is not blocked by the owner
            AND NOT EXISTS (
                SELECT 1 FROM public.blocked_users
                WHERE (blocker_id = w.user_id AND blocked_id = $2)
                   OR (blocker_id = $2 AND blocked_id = w.user_id)
            )
            -- GUARD: If profile is private, requester must be an accepted follower
            AND (
                p.is_private_account = false 
                OR EXISTS (
                    SELECT 1 FROM public.follows 
                    WHERE follower_id = $2 AND following_id = w.user_id AND status = 'accepted'
                )
            )
      )
    )
`

const getExercisesByUserIDQuery = `
  SELECT e.id, e.user_id, e.name, e.suggested_rest_seconds, e.icon, e.created_at, e.updated_at
  FROM public.exercises e
  LEFT JOIN public.profiles p ON e.user_id = p.id
  WHERE (
    -- 1. Always show system library
    e.user_id IS NULL 
    -- 2. Show the specific user's exercises ONLY IF not blocked and permitted
    OR (
      e.user_id = $1
      AND NOT EXISTS (
				SELECT 1 FROM public.blocked_users
				WHERE (blocker_id = $1 AND blocked_id = $2) 
					OR (blocker_id = $2 AND blocked_id = $1)
			)
      AND (
			p.is_private_account = false 
				OR e.user_id = $2 
				OR EXISTS (
					SELECT 1 FROM public.follows 
					WHERE follower_id = $2 
						AND following_id = $1 
						AND status = 'accepted')
				)
    )
  )
  ORDER BY (e.user_id IS NOT NULL) ASC, e.name ASC
`

const insertExerciseQuery = `
  INSERT INTO public.exercises (user_id, name, suggested_rest_seconds, icon)
  SELECT $1::uuid, $2, $3, $4
  WHERE (
      -- Regular user can create their own
      $1::uuid IS NOT NULL 
      -- Only Admin can create a system-level exercise (NULL user_id)
      OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $5::uuid)
  )
  RETURNING id, user_id, name, suggested_rest_seconds, icon, created_at, updated_at
`

const deleteExerciseByIDQuery = `
  DELETE FROM public.exercises
  WHERE id = $1
    AND (
        user_id = $2 
        OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2)
    )
`
