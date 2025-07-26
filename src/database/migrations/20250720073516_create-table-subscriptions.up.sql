-- Create ENUM types for subscription fields
DO $$ BEGIN CREATE TYPE subscription_plan_enum AS ENUM (
    'FREE',
    'BASIC',
    'PRO',
    'ENTERPRISE'
);

EXCEPTION
WHEN duplicate_object THEN null;

END $$;

DO $$ BEGIN CREATE TYPE subscription_status_enum AS ENUM (
    'ACTIVE',
    'INACTIVE',
    'EXPIRED',
    'SUSPENDED'
);

EXCEPTION
WHEN duplicate_object THEN null;

END $$;

CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY NOT NULL,
    user_id UUID NOT NULL,
    plan subscription_plan_enum NOT NULL DEFAULT 'FREE',
    status subscription_status_enum NOT NULL DEFAULT 'ACTIVE',
    points_balance INT NOT NULL DEFAULT 0,
    points_used INT NOT NULL DEFAULT 0,
    points_total INT NOT NULL DEFAULT 0,
    monthly_points INT NOT NULL DEFAULT 0,
    last_reset_date TIMESTAMP NOT NULL,
    next_reset_date TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id)
);