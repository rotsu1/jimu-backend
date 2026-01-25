package repository

const upsertFollowQuery = `
	INSERT INTO public.follows (follower_id, following_id, status)
	SELECT 
			$1,               -- follower_id
			p.id,             -- following_id
			CASE 
					WHEN p.is_private_account = true THEN 'pending' 
					ELSE 'accepted' 
			END               -- dynamic status
	FROM public.profiles p
	WHERE p.id = $2       -- Look up the person we want to follow
	ON CONFLICT (follower_id, following_id) 
	DO UPDATE SET 
			-- If it's already accepted, don't change it back to pending!
			status = CASE 
					WHEN follows.status = 'accepted' THEN 'accepted'
					ELSE EXCLUDED.status 
			END
	RETURNING follower_id, following_id, status, created_at;
`

const getFollowQuery = `
  SELECT f.follower_id, f.following_id, f.status, f.created_at
  FROM public.follows f
  WHERE f.follower_id = $1 AND f.following_id = $2
  -- The "Guard": Only return the row if NO block exists
  AND NOT EXISTS (
      SELECT 1 FROM public.blocked_users b
      WHERE (b.blocker_id = $1 AND b.blocked_id = $2)
         OR (b.blocker_id = $2 AND b.blocked_id = $1)
  )
`

const getFollowersByUserIDQuery = `
  SELECT f.follower_id, f.following_id, f.status, f.created_at
  FROM public.follows f
  WHERE f.following_id = $1 
    AND f.status = 'accepted' -- Usually, you only want to show active followers
    AND NOT EXISTS (
        SELECT 1 FROM public.blocked_users b
        WHERE (b.blocker_id = f.follower_id AND b.blocked_id = f.following_id)
           OR (b.blocker_id = f.following_id AND b.blocked_id = f.follower_id)
    )
  ORDER BY f.created_at DESC
  LIMIT $2 OFFSET $3
`

const getFollowingByUserIDQuery = `
  SELECT f.follower_id, f.following_id, f.status, f.created_at
  FROM public.follows f
  WHERE f.follower_id = $1        -- The user is the one doing the following
    AND f.status = 'accepted'      -- Only show active connections
    AND NOT EXISTS (               -- The "Double Guard" (just to be safe!)
        SELECT 1 FROM public.blocked_users b
        WHERE (b.blocker_id = f.follower_id AND b.blocked_id = f.following_id)
           OR (b.blocker_id = f.following_id AND b.blocked_id = f.follower_id)
    )
  ORDER BY f.created_at DESC
  LIMIT $2 OFFSET $3
`

const updateFollowStatusQuery = `
  UPDATE public.follows
  SET status = 'accepted'
  WHERE follower_id = $1 
    AND following_id = $2 
    AND status = 'pending'
`

const deleteFollowQuery = `
	DELETE FROM public.follows
	WHERE follower_id = $1 AND following_id = $2
`
