package repository

const getWorkoutByIDQuery = `
  SELECT 
    w.id, w.user_id, w.name, w.comment, w.started_at, w.ended_at, 
    w.duration_seconds, w.total_weight, w.likes_count, w.comments_count, 
    w.created_at, w.updated_at
  FROM public.workouts w
  JOIN public.profiles p ON w.user_id = p.id
  WHERE w.id = $1
    -- 1. Block Guard: Neither user has blocked the other
    AND NOT EXISTS (
        SELECT 1 FROM public.blocked_users b
        WHERE (b.blocker_id = w.user_id AND b.blocked_id = $2)
           OR (b.blocker_id = $2 AND b.blocked_id = w.user_id)
    )
    -- 2. Privacy Guard: Account is public OR (User is following AND accepted) OR User is viewing their own workout
    AND (
        p.is_private_account = false 
        OR w.user_id = $2
        OR EXISTS (
            SELECT 1 FROM public.follows f
            WHERE f.follower_id = $2 
              AND f.following_id = w.user_id 
              AND f.status = 'accepted'
        )
    )
`

const getWorkoutsByUserIDQuery = `
  SELECT 
    w.id, w.user_id, w.name, w.comment, w.started_at, w.ended_at, 
    w.duration_seconds, w.total_weight, w.likes_count, w.comments_count, 
    w.created_at, w.updated_at
  FROM public.workouts w
  JOIN public.profiles p ON w.user_id = p.id
  WHERE w.user_id = $1
    -- 1. Block Guard: Neither user has blocked the other
    AND NOT EXISTS (
        SELECT 1 FROM public.blocked_users b
        WHERE (b.blocker_id = w.user_id AND b.blocked_id = $2)
           OR (b.blocker_id = $2 AND b.blocked_id = w.user_id)
    )
    -- 2. Privacy Guard: Account is public OR (User is following AND accepted) OR User is viewing their own feed
    AND (
        p.is_private_account = false 
        OR w.user_id = $2
        OR EXISTS (
            SELECT 1 FROM public.follows f
            WHERE f.follower_id = $2 
              AND f.following_id = w.user_id 
              AND f.status = 'accepted'
        )
    )
  ORDER BY w.started_at DESC
  LIMIT $3 OFFSET $4
`

const insertWorkoutQuery = `
	INSERT INTO public.workouts (user_id, name, comment, started_at, ended_at, duration_seconds)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, user_id, name, comment, started_at, ended_at, duration_seconds, total_weight, likes_count, comments_count, created_at, updated_at
`

const deleteWorkoutByIDQuery = `
  DELETE FROM public.workouts
  WHERE id = $1
  AND (
      user_id = $2 
      OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2)
  )
`
