-- Note: PostgreSQL doesn't support removing enum values directly.
-- This migration would require recreating the enum type with the old values,
-- which is a complex operation that should be handled carefully in production.

-- For reference, to remove an enum value, you would need to:
-- 1. Create a new enum type without the unwanted value
-- 2. Update all tables to use the new enum type
-- 3. Drop the old enum type
-- 4. Rename the new enum type to the original name

-- This is typically not recommended for production environments.
-- Consider keeping the enum value but ensuring it's not used in the application. 