package repository

const getCommentByIDQuery = `
  SELECT c.id, c.user_id, c.workout_id, c.parent_id, c.content, c.likes_count, c.created_at
  FROM public.comments c
  JOIN public.workouts w ON c.workout_id = w.id
  JOIN public.profiles p ON w.user_id = p.id
  WHERE c.id = $1
    -- 1. Block Guard: Neither the viewer nor the workout owner have blocked each other
    AND NOT EXISTS (
        SELECT 1 FROM public.blocked_users b
        WHERE (b.blocker_id = w.user_id AND b.blocked_id = $2)
           OR (b.blocker_id = $2 AND b.blocked_id = w.user_id)
    )
    -- 2. Privacy Guard: Public account OR viewing own workout OR accepted follower
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

const getCommentsByWorkoutIDQuery = `
  SELECT c.id, c.user_id, c.workout_id, c.parent_id, c.content, c.likes_count, c.created_at
  FROM public.comments c
  JOIN public.workouts w ON c.workout_id = w.id
  JOIN public.profiles p ON w.user_id = p.id
  WHERE c.workout_id = $1 AND c.parent_id IS NULL
    -- 1. Block Guard: Neither the viewer nor the workout owner have blocked each other
    AND NOT EXISTS (
        SELECT 1 FROM public.blocked_users b
        WHERE (b.blocker_id = w.user_id AND b.blocked_id = $2)
           OR (b.blocker_id = $2 AND b.blocked_id = w.user_id)
    )
    -- 2. Privacy Guard: Public account OR viewing own workout OR accepted follower
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
    -- 3. Ghost Filter: Hide comments from blocked users
    AND NOT EXISTS (
        SELECT 1 FROM public.blocked_users b
        WHERE (b.blocker_id = c.user_id AND b.blocked_id = $2)
           OR (b.blocker_id = $2 AND b.blocked_id = c.user_id)
    )
  ORDER BY c.created_at ASC
  LIMIT $3 OFFSET $4
`

const getRepliesByCommentIDQuery = `
  SELECT c.id, c.user_id, c.workout_id, c.parent_id, c.content, c.likes_count, c.created_at
  FROM public.comments c
  JOIN public.comments parent ON c.parent_id = parent.id
  JOIN public.workouts w ON parent.workout_id = w.id
  JOIN public.profiles p ON w.user_id = p.id
  WHERE c.parent_id = $1
    -- 1. Block Guard: Neither the viewer nor the workout owner have blocked each other
    AND NOT EXISTS (
        SELECT 1 FROM public.blocked_users b
        WHERE (b.blocker_id = w.user_id AND b.blocked_id = $2)
           OR (b.blocker_id = $2 AND b.blocked_id = w.user_id)
    )
    -- 2. Privacy Guard: Public account OR viewing own workout OR accepted follower
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
    -- 3. Ghost Filter: Hide replies from blocked users
    AND NOT EXISTS (
        SELECT 1 FROM public.blocked_users b
        WHERE (b.blocker_id = c.user_id AND b.blocked_id = $2)
           OR (b.blocker_id = $2 AND b.blocked_id = c.user_id)
    )
  ORDER BY c.created_at ASC
`

const insertCommentQuery = `
  INSERT INTO public.comments (user_id, workout_id, parent_id, content)
  SELECT $1, w.id, $3, $4
  FROM public.workouts w
  JOIN public.profiles p ON w.user_id = p.id
  WHERE w.id = $2
    -- 1. Block Guard: Neither the commenter nor the workout owner have blocked each other
    AND NOT EXISTS (
        SELECT 1 FROM public.blocked_users b
        WHERE (b.blocker_id = w.user_id AND b.blocked_id = $1)
           OR (b.blocker_id = $1 AND b.blocked_id = w.user_id)
    )
    -- 2. Privacy Guard: Public account OR commenting on own workout OR accepted follower
    AND (
        p.is_private_account = false 
        OR w.user_id = $1
        OR EXISTS (
            SELECT 1 FROM public.follows f
            WHERE f.follower_id = $1 
              AND f.following_id = w.user_id 
              AND f.status = 'accepted'
        )
    )
  RETURNING id, user_id, workout_id, parent_id, content, likes_count, created_at
`

const deleteCommentByIDQuery = `
  DELETE FROM public.comments
  WHERE id = $1 
  AND user_id = $2
  AND workout_id = $3
`
