package repository

const getUserSessionByIDQuery = `
	SELECT id, user_id, refresh_token, user_agent, client_ip, is_revoked, expires_at, created_at, updated_at
	FROM public.user_sessions
	WHERE id = $1
    -- Guard: Owner or Admin
    AND (
        user_id = $2
        OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2)
    )
`

// Since this will be queried when the token is expired, we cannot add a user_id guard here.
// Instead, the refresh token is the guard itself.
const getUserSessionByRefreshTokenQuery = `
	SELECT id, user_id, refresh_token, user_agent, client_ip, is_revoked, expires_at, created_at, updated_at
	FROM public.user_sessions
	WHERE refresh_token = $1 AND is_revoked = false AND expires_at > NOW()
`

const getUserSessionsByUserIDQuery = `
  SELECT id, user_id, refresh_token, user_agent, client_ip, is_revoked, expires_at, created_at, updated_at
  FROM public.user_sessions
  WHERE user_id = $1
    AND (
      user_id = $2 
      OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2)
    )
  ORDER BY created_at DESC
`

const insertUserSessionQuery = `
	INSERT INTO public.user_sessions (user_id, refresh_token, user_agent, client_ip, expires_at)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, user_id, refresh_token, user_agent, client_ip, is_revoked, expires_at, created_at, updated_at
`

const revokeUserSessionQuery = `
	UPDATE public.user_sessions
	SET is_revoked = true
	WHERE id = $1
    -- Guard: Owner or Admin
    AND (
        user_id = $2
        OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2)
    )
`

const revokeUserSessionsByUserIDQuery = `
	UPDATE public.user_sessions
	SET is_revoked = true
	WHERE user_id = $1 AND is_revoked = false
    -- Guard: Viewer is Owner ($2) or Admin ($2)
    AND (
        $1 = $2
        OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2)
    )
`

const deleteExpiredUserSessionsQuery = `
	DELETE FROM public.user_sessions
	WHERE expires_at < NOW()
    -- Admin maintenance task, usually no viewer context or system context.
    -- Typically run by a background job.
`

const deleteUserSessionByIDQuery = `
	DELETE FROM public.user_sessions
	WHERE id = $1
    -- Guard: Owner or Admin
    AND (
        user_id = $2
        OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2)
    )
`
