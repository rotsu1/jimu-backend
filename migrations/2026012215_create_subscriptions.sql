-- +migrate Up
CREATE TABLE IF NOT EXISTS public.subscriptions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL UNIQUE REFERENCES public.profiles(id),
    original_transaction_id text NOT NULL UNIQUE, 
    product_id text NOT NULL,
    status text NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    environment text NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON public.subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_original_transaction_id ON public.subscriptions(original_transaction_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON public.subscriptions(status);

-- +migrate Down
DROP TABLE IF EXISTS public.subscriptions;