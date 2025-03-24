CREATE TABLE shopify_sessions (
    id VARCHAR(255) PRIMARY KEY NOT NULL,
    access_token TEXT,
    user_id UUID NOT NULL,
    expires TIMESTAMP,
    is_online BOOLEAN DEFAULT FALSE NOT NULL,
    scope VARCHAR(255),
    state TEXT,
    api_key VARCHAR(255) NOT NULL,
    shop VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_shopify_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);