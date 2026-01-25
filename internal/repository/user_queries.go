package repository

const getUserIdentityByProviderQuery = `
			SELECT id, user_id, provider_name, provider_user_id, provider_email, last_sign_in_at, created_at, updated_at
			FROM user_identities 
			WHERE provider_name = $1 
			AND provider_user_id = $2
`

const insertUserIdentityQuery = `
			INSERT INTO user_identities (provider_name, provider_user_id, provider_email)
			VALUES ('google', $1, $2)
			RETURNING user_id;
`

const updateUserIdentityQuery = `
			UPDATE user_identities 
			SET last_sign_in_at = now(), provider_email = $2
			WHERE provider_name = 'google' AND provider_user_id = $1
`

const getProfileQuery = `
			SELECT 
			p.id,
			p.username,
			p.display_name,
			p.bio,

			CASE 
					WHEN p.id = $1 OR p.is_private_account = false OR f.status = 'accepted' 
					THEN p.location 
					ELSE '' 
			END AS location,

			CASE 
					WHEN p.id = $1 OR p.is_private_account = false OR f.status = 'accepted' 
					THEN p.birth_date 
					ELSE '0001-01-01' -- Go's time.Time zero value equivalent
			END AS birth_date,

			p.avatar_url,
			p.subscription_plan,
			p.is_private_account,

			CASE 
					WHEN p.id = $1 OR p.is_private_account = false OR f.status = 'accepted' 
					THEN p.last_worked_out_at 
					ELSE '0001-01-01' -- Go's time.Time zero value equivalent
			END AS last_worked_out_at,

			CASE 
					WHEN p.id = $1 OR p.is_private_account = false OR f.status = 'accepted' 
					THEN p.total_workouts 
					ELSE 0 
			END AS total_workouts,

			CASE 
					WHEN p.id = $1 OR p.is_private_account = false OR f.status = 'accepted' 
					THEN p.current_streak 
					ELSE 0 
			END AS current_streak,

			CASE 
					WHEN p.id = $1 OR p.is_private_account = false OR f.status = 'accepted' 
					THEN p.total_weight 
					ELSE 0 
			END AS total_weight,

			p.created_at,

			CASE
					WHEN p.id = $1 OR p.is_private_account = false OR f.status = 'accepted' 
					THEN p.updated_at 
					ELSE '0001-01-01' -- Go's time.Time zero value equivalent
			END AS updated_at

			FROM public.profiles p
			-- Check if the viewer is following the target
			LEFT JOIN public.follows f 
			ON f.following_id = p.id AND f.follower_id = $1
			-- Check if the target is not blocked (either the target blocked the viewer or the viewer blocked the target)
			LEFT JOIN public.blocks b 
			ON (b.blocker_id = p.id AND b.blocked_id = $1) 
			OR (b.blocker_id = $1 AND b.blocked_id = p.id)

			WHERE p.id = $2
			-- If there is any block relationship, return no rows (treat it as non-existent)
			AND b.blocker_id IS NULL;
`
