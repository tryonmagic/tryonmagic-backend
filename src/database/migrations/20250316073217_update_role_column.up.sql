-- Create the ENUM type for role
DO $$ BEGIN
    CREATE TYPE role_enum AS ENUM ('USER', 'ADMIN', 'INTERNAL_SERVICE');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Alter the role column to use the ENUM type
ALTER TABLE users
ALTER COLUMN role DROP DEFAULT,
ALTER COLUMN role TYPE role_enum USING role::role_enum,
ALTER COLUMN role SET DEFAULT 'USER';