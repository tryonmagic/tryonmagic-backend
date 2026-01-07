-- Drop indexes
DROP INDEX IF EXISTS idx_users_shop_shopify_user_id;

DROP INDEX IF EXISTS idx_user_shop_shopify_user_id;

DROP INDEX IF EXISTS idx_users_email_unique_non_shopify;

-- Restore unique constraint on email
ALTER TABLE
    users
ADD
    CONSTRAINT users_email_key UNIQUE (email);

-- Remove shop and shopify_user_id columns
ALTER TABLE
    users DROP COLUMN IF EXISTS shop,
    DROP COLUMN IF EXISTS shopify_user_id;