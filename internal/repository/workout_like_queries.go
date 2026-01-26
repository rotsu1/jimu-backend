package repository

const insertWorkoutLikeQuery = `
  INSERT INTO public.workout_likes (user_id, workout_id)
  SELECT $1, w.id
  FROM public.workouts w
  JOIN public.profiles p ON w.user_id = p.id
  WHERE w.id = $2
    -- 1. Block Guard: Neither user has blocked the other
    AND NOT EXISTS (
        SELECT 1 FROM public.blocked_users b
        WHERE (b.blocker_id = w.user_id AND b.blocked_id = $1)
           OR (b.blocker_id = $1 AND b.blocked_id = w.user_id)
    )
    -- 2. Privacy Guard: Public account OR (Follower and accepted) OR Own workout
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
  ON CONFLICT (user_id, workout_id) DO NOTHING
  RETURNING user_id, workout_id, created_at
`

const getWorkoutLikeQuery = `
  SELECT 
    l.user_id, 
    l.workout_id, 
    l.created_at
  FROM public.workout_likes l
  JOIN public.workouts w ON l.workout_id = w.id
  JOIN public.profiles owner_p ON w.user_id = owner_p.id
  WHERE l.user_id = $1 AND l.workout_id = $2
    -- 1. Block Guard: Neither the viewer nor the owner have blocked each other
    AND NOT EXISTS (
        SELECT 1 FROM public.blocked_users b
        WHERE (b.blocker_id = w.user_id AND b.blocked_id = $1)
           OR (b.blocker_id = $1 AND b.blocked_id = w.user_id)
    )
    -- 2. Privacy Guard: Public account OR viewing own workout OR accepted follower
    AND (
        owner_p.is_private_account = false 
        OR w.user_id = $1
        OR EXISTS (
            SELECT 1 FROM public.follows f
            WHERE f.follower_id = $1 
              AND f.following_id = w.user_id 
              AND f.status = 'accepted'
        )
    )
`

const getLikesByWorkoutIDQuery = `
  SELECT 
    l.user_id, l.workout_id, l.created_at,
    p.username, p.avatar_url -- You usually want to show who they are!
  FROM public.workout_likes l
  JOIN public.workouts w ON l.workout_id = w.id
  JOIN public.profiles owner_p ON w.user_id = owner_p.id
  JOIN public.profiles p ON l.user_id = p.id
  WHERE l.workout_id = $1
    -- 1. Can the Viewer see the Workout at all? (Owner Block/Privacy)
    AND NOT EXISTS (
        SELECT 1 FROM public.blocked_users b
        WHERE (b.blocker_id = w.user_id AND b.blocked_id = $2)
           OR (b.blocker_id = $2 AND b.blocked_id = w.user_id)
    )
    AND (owner_p.is_private_account = false OR w.user_id = $2 OR EXISTS (
        SELECT 1 FROM public.follows f 
        WHERE f.follower_id = $2 AND f.following_id = w.user_id AND f.status = 'accepted'
    ))
    -- 2. "The Ghost Filter": Hide individual likers who have a block with the viewer
    AND NOT EXISTS (
        SELECT 1 FROM public.blocked_users b
        WHERE (b.blocker_id = l.user_id AND b.blocked_id = $2)
           OR (b.blocker_id = $2 AND b.blocked_id = l.user_id)
    )
  ORDER BY l.created_at DESC
  LIMIT $3 OFFSET $4
`

const deleteWorkoutLikeQuery = `
	DELETE FROM public.workout_likes
	WHERE user_id = $1 AND workout_id = $2
`

const isWorkoutLikedQuery = `
  SELECT EXISTS (
    SELECT 1 
    FROM public.workout_likes l
    JOIN public.workouts w ON l.workout_id = w.id
    JOIN public.profiles p ON w.user_id = p.id
    WHERE l.user_id = $1 AND l.workout_id = $2
      -- Block Guard
      AND NOT EXISTS (
          SELECT 1 FROM public.blocked_users b
          WHERE (b.blocker_id = w.user_id AND b.blocked_id = $1)
             OR (b.blocker_id = $1 AND b.blocked_id = w.user_id)
      )
      -- Privacy Guard
      AND (
          p.is_private_account = false 
          OR w.user_id = $1
          OR EXISTS (
              SELECT 1 FROM public.follows f
              WHERE f.follower_id = $1 AND f.following_id = w.user_id AND f.status = 'accepted'
          )
      )
  )
`
