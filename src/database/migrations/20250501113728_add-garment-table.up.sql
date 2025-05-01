
-- Create the ENUM type for role
DO $$ BEGIN
    CREATE TYPE garment_outfit_type_enum AS ENUM ('TOP', 'BOTTOM', 'FULL_BODY');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE garments (
    id UUID PRIMARY KEY NOT NULL,
    outfit_type garment_outfit_type_enum NOT NULL,
    color VARCHAR(50) NOT NULL,
    image_url VARCHAR(255) NOT NULL,
    tags VARCHAR(255) NOT NULL,
    user_id UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id)
);