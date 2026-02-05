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

const getTimelineWorkoutsQuery = `
  SELECT 
    w.id, 
    w.user_id, 
    p.username, 
    p.avatar_url, 
    w.name, 
    w.comment, 
    w.started_at, 
    w.ended_at, 
    w.total_weight, 
    w.likes_count, 
    w.comments_count, 
    w.updated_at,
    
    -- 1. EXERCISES (Nested with Sets)
    COALESCE(
      (
        SELECT json_agg(
          json_build_object(
            'id', we.id,
            'exercise_id', we.exercise_id,
            'name', e.name, 
            'order_index', we.order_index,
            'sets', COALESCE(
              (
                SELECT json_agg(
                  json_build_object(
                    'id', ws.id,
                    'weight', ws.weight,
                    'reps', ws.reps,
                    'order_index', ws.order_index
                  ) ORDER BY ws.order_index ASC
                )
                FROM public.workout_sets ws
                WHERE ws.workout_exercise_id = we.id
              ), '[]'::json
            )
          ) ORDER BY we.order_index ASC
        )
        FROM public.workout_exercises we
        JOIN public.exercises e ON we.exercise_id = e.id
        WHERE we.workout_id = w.id
      ), '[]'::json
    ) AS exercises,

    -- 2. COMMENTS (Joined with Profile)
    COALESCE(
      (
        SELECT json_agg(
          json_build_object(
            'id', wc.id,
            'user_id', wc.user_id,
            'content', wc.content,
            'likes_count', wc.likes_count,
            'created_at', wc.created_at,
            'username', cp.username,      -- Commenter's username
            'avatar_url', cp.avatar_url,  -- Commenter's avatar
            'comments', '[]'::json        -- Empty array for child comments (Lazy Load these!)
          ) ORDER BY wc.created_at ASC
        )
        FROM public.comments wc
        JOIN public.profiles cp ON wc.user_id = cp.id
        WHERE wc.workout_id = w.id
      ), '[]'::json
    ) AS comments,

    -- 3. IMAGES
    COALESCE(
      (
        SELECT json_agg(
          json_build_object(
            'id', wi.id,
            'storage_path', wi.storage_path,
            'display_order', wi.display_order
          ) ORDER BY wi.display_order ASC
        )
        FROM public.workout_images wi
        WHERE wi.workout_id = w.id
      ), '[]'::json
    ) AS images

  FROM public.workouts w
  JOIN public.profiles p ON w.user_id = p.id
  WHERE w.user_id = $1
    -- Block & Privacy Logic
    AND NOT EXISTS (
        SELECT 1 FROM public.blocked_users b
        WHERE (b.blocker_id = w.user_id AND b.blocked_id = $2)
           OR (b.blocker_id = $2 AND b.blocked_id = w.user_id)
    )
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
