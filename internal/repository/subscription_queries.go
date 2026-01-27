package repository

const getSubscriptionByUserIDQuery = `
  SELECT id, user_id, original_transaction_id, product_id, status, expires_at, environment, created_at, updated_at
  FROM public.subscriptions
  WHERE user_id = $1
    AND (
      user_id = $2 
      OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2)
    )
`

const getSubscriptionByTransactionIDQuery = `
  SELECT id, user_id, original_transaction_id, product_id, status, expires_at, environment, created_at, updated_at
  FROM public.subscriptions
  WHERE original_transaction_id = $1
    AND (
      -- 1. The requester is the owner of this subscription
      user_id = $2 
      -- 2. The requester is a system administrator
      OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $2)
    )
`

const upsertSubscriptionQuery = `
	INSERT INTO public.subscriptions (user_id, original_transaction_id, product_id, status, expires_at, environment)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (user_id) 
	DO UPDATE SET 
		original_transaction_id = EXCLUDED.original_transaction_id,
		product_id = EXCLUDED.product_id,
		status = EXCLUDED.status,
		expires_at = EXCLUDED.expires_at,
		environment = EXCLUDED.environment
	RETURNING id, user_id, original_transaction_id, product_id, status, expires_at, environment, created_at, updated_at
`

const deleteSubscriptionByUserIDQuery = `
	DELETE FROM public.subscriptions
	WHERE user_id = $1
`
