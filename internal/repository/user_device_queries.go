package repository

const getUserDeviceByIDQuery = `
	SELECT id, user_id, fcm_token, device_type, created_at, updated_at
	FROM public.user_devices
	WHERE id = $1
  	AND (user_id = $2 OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2))
`

const getUserDevicesByUserIDQuery = `
	SELECT id, user_id, fcm_token, device_type, created_at, updated_at
	FROM public.user_devices
	WHERE user_id = $1
		AND (user_id = $2 OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2))
	ORDER BY updated_at DESC
`

const getUserDeviceByFCMTokenQuery = `
	SELECT id, user_id, fcm_token, device_type, created_at, updated_at
	FROM public.user_devices
	WHERE fcm_token = $1
  	AND (user_id = $2 OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2))
`

const upsertUserDeviceQuery = `
  INSERT INTO public.user_devices (user_id, fcm_token, device_type)
  SELECT $1::uuid, $2, $3
  WHERE (
      -- Guard: You can only register a device for YOURSELF
      $1::uuid = $4::uuid
      -- Or you are an admin (unlikely for devices, but good for consistency)
      OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $4::uuid)
  )
  ON CONFLICT (fcm_token) 
  DO UPDATE SET 
    user_id = EXCLUDED.user_id,
    device_type = EXCLUDED.device_type,
    updated_at = now()
  RETURNING id, user_id, fcm_token, device_type, created_at, updated_at
`

const deleteUserDeviceByIDQuery = `
DELETE FROM public.user_devices
	WHERE id = $1
		AND (user_id = $2 OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2))
`

const deleteUserDevicesByUserIDQuery = `
	DELETE FROM public.user_devices
	WHERE user_id = $1
		AND (user_id = $2 OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2))
`
