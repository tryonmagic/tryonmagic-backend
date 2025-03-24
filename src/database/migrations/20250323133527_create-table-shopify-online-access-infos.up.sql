CREATE TABLE shopify_online_access_infos (
    id UUID PRIMARY KEY NOT NULL,
    shopify_session_id VARCHAR(255) UNIQUE NOT NULL,
    expires_in INT NOT NULL,
    associated_user_scope TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_online_access_session FOREIGN KEY (shopify_session_id) REFERENCES shopify_sessions(id) ON DELETE CASCADE
);