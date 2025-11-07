-- Remove the foreign key constraints added in the up migration
ALTER TABLE
    shopify_associated_users DROP CONSTRAINT IF EXISTS fk_shopify_associated_users_access_info;

ALTER TABLE
    shopify_online_access_infos DROP CONSTRAINT IF EXISTS fk_shopify_online_access_infos_session;