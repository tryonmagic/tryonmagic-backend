-- Add shop and shopify_user_id columns to users table
ALTER TABLE
    users
ADD
    COLUMN shop VARCHAR(255),
ADD
    COLUMN shopify_user_id BIGINT;

-- Drop the unique constraint on email (we'll replace it with conditional unique indexes)
ALTER TABLE
    users DROP CONSTRAINT IF EXISTS users_email_key;

-- Create partial unique index on email for non-Shopify users (where shop IS NULL)
-- This ensures email uniqueness for regular users
CREATE UNIQUE INDEX idx_users_email_unique_non_shopify ON users(email)
WHERE
    shop IS NULL;

-- Create composite unique index on (shop, shopify_user_id) for Shopify users
-- This ensures each shop+shopify_user_id combination is unique
CREATE UNIQUE INDEX idx_user_shop_shopify_user_id ON users(shop, shopify_user_id)
WHERE
    shop IS NOT NULL
    AND shopify_user_id IS NOT NULL;

-- Create index for faster lookups
CREATE INDEX idx_users_shop_shopify_user_id ON users(shop, shopify_user_id);