-- Revert the role column to VARCHAR
ALTER TABLE users
ALTER COLUMN role DROP DEFAULT,
ALTER COLUMN role TYPE VARCHAR(255) USING role::VARCHAR(255);

-- Drop the ENUM type
DROP TYPE IF EXISTS role_enum;