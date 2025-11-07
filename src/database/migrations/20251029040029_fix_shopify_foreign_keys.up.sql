-- Add proper foreign key constraints with cascade delete for Shopify session tables
-- This will ensure that when a ShopifySession is deleted, related records are automatically deleted
-- First, drop existing foreign key constraints if they exist
ALTER TABLE
    shopify_online_access_infos DROP CONSTRAINT IF EXISTS fk_shopify_online_access_infos_session;

ALTER TABLE
    shopify_associated_users DROP CONSTRAINT IF EXISTS fk_shopify_associated_users_access_info;

-- Add foreign key constraint for shopify_online_access_infos with CASCADE DELETE
ALTER TABLE
    shopify_online_access_infos
ADD
    CONSTRAINT fk_shopify_online_access_infos_session FOREIGN KEY (shopify_session_id) REFERENCES shopify_sessions(id) ON DELETE CASCADE;

-- Add foreign key constraint for shopify_associated_users with CASCADE DELETE
ALTER TABLE
    shopify_associated_users
ADD
    CONSTRAINT fk_shopify_associated_users_access_info FOREIGN KEY (online_access_info_id) REFERENCES shopify_online_access_infos(id) ON DELETE CASCADE;