package repository

const insertBlockedUserQuery = `
	INSERT INTO public.blocked_users (blocker_id, blocked_id)
	VALUES ($1, $2)
	ON CONFLICT (blocker_id, blocked_id) DO NOTHING
	RETURNING blocker_id, blocked_id, created_at
`

const getBlockedUserQuery = `
	SELECT blocker_id, blocked_id, created_at
	FROM public.blocked_users
	WHERE blocker_id = $1 AND blocked_id = $2
`

const getBlockedUsersByBlockerIDQuery = `
	SELECT blocker_id, blocked_id, created_at
	FROM public.blocked_users
	WHERE blocker_id = $1
	ORDER BY created_at DESC
`

const deleteBlockedUserQuery = `
	DELETE FROM public.blocked_users
	WHERE blocker_id = $1 AND blocked_id = $2
`

const isBlockedQuery = `
	SELECT EXISTS(
		SELECT 1 FROM public.blocked_users
		WHERE (blocker_id = $1 AND blocked_id = $2)
		   OR (blocker_id = $2 AND blocked_id = $1)
	)
`
