
-- Create the ENUM type for role
DO $$ BEGIN
    CREATE TYPE ai_model_author_enum AS ENUM ('USER', 'SYSTEM');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

ALTER TYPE task_type_enum ADD VALUE 'GENERATE_AI_MODEL';
-- Create the ai_models table
CREATE TABLE ai_models (
    id UUID PRIMARY KEY NOT NULL,
    ai_model_author ai_model_author_enum NOT NULL,
    model_title VARCHAR(100) NOT NULL,
    prompt TEXT NOT NULL,
    negative_prompt TEXT,
    gen_ai_name VARCHAR(100) NOT NULL,
    model_category VARCHAR(100) NOT NULL,
    model_location VARCHAR(100) NOT NULL,
    ethnicity VARCHAR(100) NOT NULL,
    skin_tone VARCHAR(100),
    gender VARCHAR(50) NOT NULL,
    age_range VARCHAR(50) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    style_tags VARCHAR(255),
    resolution JSONB,
    aspect_ratio VARCHAR(50),
    override_parameters BOOLEAN DEFAULT FALSE,
    user_id UUID,
    task_id UUID,
    model_image_url VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT fk_task FOREIGN KEY (task_id) REFERENCES tasks (id)
);