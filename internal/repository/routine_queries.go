package repository

const getRoutineByIDQuery = `
  SELECT r.id, r.user_id, r.name, r.created_at, r.updated_at
  FROM public.routines r
  JOIN public.profiles p ON r.user_id = p.id
  WHERE r.id = $1
    -- 1. Block Guard: Neither user has blocked the other
    AND NOT EXISTS (
        SELECT 1 FROM public.blocked_users b
        WHERE (b.blocker_id = r.user_id AND b.blocked_id = $2)
           OR (b.blocker_id = $2 AND b.blocked_id = r.user_id)
    )
    -- 2. Privacy Guard: Public account OR (User is following AND accepted) OR User is viewing their own routine
    AND (
        p.is_private_account = false 
        OR r.user_id = $2
        OR EXISTS (
            SELECT 1 FROM public.follows f
            WHERE f.follower_id = $2 
              AND f.following_id = r.user_id 
              AND f.status = 'accepted'
        )
    )
`

const getRoutinesByUserIDQuery = `
  SELECT r.id, r.user_id, r.name, r.created_at, r.updated_at
  FROM public.routines r
  JOIN public.profiles p ON r.user_id = p.id
  WHERE r.user_id = $1
    -- 1. Block Guard
    AND NOT EXISTS (
        SELECT 1 FROM public.blocked_users b
        WHERE (b.blocker_id = r.user_id AND b.blocked_id = $2)
           OR (b.blocker_id = $2 AND b.blocked_id = r.user_id)
    )
    -- 2. Privacy Guard
    AND (
        p.is_private_account = false 
        OR r.user_id = $2
        OR EXISTS (
            SELECT 1 FROM public.follows f
            WHERE f.follower_id = $2 
              AND f.following_id = r.user_id 
              AND f.status = 'accepted'
        )
    )
  ORDER BY r.name ASC
`

const insertRoutineQuery = `
	INSERT INTO public.routines (user_id, name)
	VALUES ($1, $2)
	RETURNING id, user_id, name, created_at, updated_at
`

const updateRoutineQuery = `
  UPDATE public.routines
  SET name = COALESCE($2, name), updated_at = NOW()
  WHERE id = $3 AND user_id = $1
`

const deleteRoutineByIDQuery = `
  DELETE FROM public.routines
  WHERE id = $1 
  AND (
      user_id = $2 
      OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2)
  )
`
