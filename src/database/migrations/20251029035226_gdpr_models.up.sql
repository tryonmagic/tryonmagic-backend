CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- GDPR Request table
CREATE TABLE gdpr_requests(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop VARCHAR(255) NOT NULL,
    customer_id VARCHAR(255) NOT NULL,
    customer_email VARCHAR(255) NOT NULL,
    request_type VARCHAR(50) NOT NULL CHECK (
        request_type IN ('data_request', 'redact', 'shop_redact')
    ),
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (
        status IN ('pending', 'processing', 'completed', 'failed')
    ),
    payload JSONB,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    processed_at TIMESTAMP
);

-- Create indexes for GDPR requests
CREATE INDEX idx_gdpr_requests_shop ON gdpr_requests(shop);

CREATE INDEX idx_gdpr_requests_customer_id ON gdpr_requests(customer_id);

CREATE INDEX idx_gdpr_requests_status ON gdpr_requests(status);

CREATE INDEX idx_gdpr_requests_created_at ON gdpr_requests(created_at);

-- User Preferences table
CREATE TABLE user_preferences(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop VARCHAR(255) NOT NULL,
    customer_id VARCHAR(255) NOT NULL,
    preferences JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Create indexes for user preferences
CREATE INDEX idx_user_preferences_shop ON user_preferences(shop);

CREATE INDEX idx_user_preferences_customer_id ON user_preferences(customer_id);

-- Generated Images table
CREATE TABLE generated_images(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop VARCHAR(255) NOT NULL,
    customer_id VARCHAR(255) NOT NULL,
    image_url VARCHAR(500) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Create indexes for generated images
CREATE INDEX idx_generated_images_shop ON generated_images(shop);

CREATE INDEX idx_generated_images_customer_id ON generated_images(customer_id);

CREATE INDEX idx_generated_images_created_at ON generated_images(created_at);

-- Usage History table
CREATE TABLE usage_history(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop VARCHAR(255) NOT NULL,
    customer_id VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Create indexes for usage history
CREATE INDEX idx_usage_history_shop ON usage_history(shop);

CREATE INDEX idx_usage_history_customer_id ON usage_history(customer_id);

CREATE INDEX idx_usage_history_created_at ON usage_history(created_at);

-- Shop Settings table
CREATE TABLE shop_settings(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop VARCHAR(255) NOT NULL UNIQUE,
    settings JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Create index for shop settings
CREATE INDEX idx_shop_settings_shop ON shop_settings(shop);

-- Shop Deletions audit log table
CREATE TABLE shop_deletions(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop VARCHAR(255) NOT NULL,
    deleted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    reason VARCHAR(255) NOT NULL,
    data_deleted INTEGER DEFAULT 0 NOT NULL
);

-- Create index for shop deletions
CREATE INDEX idx_shop_deletions_shop ON shop_deletions(shop);

CREATE INDEX idx_shop_deletions_deleted_at ON shop_deletions(deleted_at);