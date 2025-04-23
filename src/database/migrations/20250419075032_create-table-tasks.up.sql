-- Create the ENUM type for role
DO $$ BEGIN
    CREATE TYPE task_type_enum AS ENUM ('REMOVE_BACKGROUND', 'GENERATE_IMAGE', 'REPLACE_BACKGROUND');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE task_status_enum AS ENUM ('PENDING', 'QUEUED', 'PROCESSING', 'SUCCEEDED', 'FAILED');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;



-- Alter the role column to use the ENUM type
ALTER TABLE users
ALTER COLUMN role DROP DEFAULT,
ALTER COLUMN role TYPE role_enum USING role::role_enum,
ALTER COLUMN role SET DEFAULT 'USER';

CREATE TABLE tasks (
    id UUID PRIMARY KEY NOT NULL,
    type task_type_enum NOT NULL,
    status task_status_enum NOT NULL,
    input_data JSONB NOT NULL,
    result_data JSONB,
    error TEXT,
    retry_count INT DEFAULT 0,
    user_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id)
);