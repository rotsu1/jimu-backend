package repository

const insertCommentLikeQuery = `
  INSERT INTO public.comment_likes (user_id, comment_id)
  SELECT $1, c.id
  FROM public.comments c
  JOIN public.workouts w ON c.workout_id = w.id
  JOIN public.profiles p ON w.user_id = p.id
  WHERE c.id = $2
    -- 1. Block Guard: Neither the liker nor the workout owner have blocked each other
    AND NOT EXISTS (
        SELECT 1 FROM public.blocked_users b
        WHERE (b.blocker_id = w.user_id AND b.blocked_id = $1)
           OR (b.blocker_id = $1 AND b.blocked_id = w.user_id)
    )
    -- 2. Privacy Guard: Public account OR own workout OR accepted follower
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
  ON CONFLICT (user_id, comment_id) DO NOTHING
  RETURNING user_id, comment_id, created_at
`

const getCommentLikeQuery = `
  SELECT l.user_id, l.comment_id, l.created_at
  FROM public.comment_likes l
  JOIN public.comments c ON l.comment_id = c.id
  JOIN public.workouts w ON c.workout_id = w.id
  JOIN public.profiles p ON w.user_id = p.id
  WHERE l.user_id = $1 AND l.comment_id = $2
    -- 1. Block Guard: Neither the viewer nor the workout owner have blocked each other
    AND NOT EXISTS (
        SELECT 1 FROM public.blocked_users b
        WHERE (b.blocker_id = w.user_id AND b.blocked_id = $1)
           OR (b.blocker_id = $1 AND b.blocked_id = w.user_id)
    )
    -- 2. Privacy Guard: Public account OR viewing own workout OR accepted follower
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
`

const getCommentLikesByCommentIDQuery = `

  SELECT 
    l.user_id, 
    l.comment_id, 
    l.created_at,
    liker_p.username,      -- The person who liked it
    liker_p.avatar_url     -- The person who liked it
  FROM public.comment_likes l
  JOIN public.profiles liker_p ON l.user_id = liker_p.id -- Join the Liker
  JOIN public.comments c ON l.comment_id = c.id
  JOIN public.workouts w ON c.workout_id = w.id
  JOIN public.profiles owner_p ON w.user_id = owner_p.id -- Join the Workout Owner
  WHERE l.comment_id = $1
    -- Admin Bypass: If viewer is admin, skip all social guards
    AND (
      EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2)
      OR (
        -- 1. Context Guard (Can I see the Workout Owner's content?)
        NOT EXISTS (
            SELECT 1 FROM public.blocked_users b
            WHERE (b.blocker_id = w.user_id AND b.blocked_id = $2)
               OR (b.blocker_id = $2 AND b.blocked_id = w.user_id)
        )
        AND (owner_p.is_private_account = false OR w.user_id = $2 OR EXISTS (
            SELECT 1 FROM public.follows f
            WHERE f.follower_id = $2 AND f.following_id = w.user_id AND f.status = 'accepted'
        ))
        -- 2. Ghost Filter (Block between Viewer and the specific Liker)
        AND NOT EXISTS (
            SELECT 1 FROM public.blocked_users b
            WHERE (b.blocker_id = l.user_id AND b.blocked_id = $2)
               OR (b.blocker_id = $2 AND b.blocked_id = l.user_id)
        )
      )
    )
    ORDER BY l.created_at DESC
    LIMIT $3 OFFSET $4
`

const deleteCommentLikeQuery = `
	DELETE FROM public.comment_likes
	WHERE user_id = $1 AND comment_id = $2
`

const isCommentLikedQuery = `
  SELECT EXISTS(
    SELECT 1 FROM public.comment_likes l
    JOIN public.comments c ON l.comment_id = c.id
    JOIN public.workouts w ON c.workout_id = w.id
    JOIN public.profiles p ON w.user_id = p.id
    WHERE l.user_id = $1 AND l.comment_id = $2
      -- 1. Block Guard: Neither the viewer nor the workout owner have blocked each other
      AND NOT EXISTS (
          SELECT 1 FROM public.blocked_users b
          WHERE (b.blocker_id = w.user_id AND b.blocked_id = $1)
             OR (b.blocker_id = $1 AND b.blocked_id = w.user_id)
      )
      -- 2. Privacy Guard: Public account OR viewing own workout OR accepted follower
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
  )
`
