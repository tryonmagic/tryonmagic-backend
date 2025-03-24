CREATE TABLE shopify_associated_users (
    id UUID PRIMARY KEY NOT NULL,
    online_access_info_id UUID UNIQUE,
    shopify_user_id BIGINT NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    account_owner BOOLEAN NOT NULL,
    locale VARCHAR(20) NOT NULL,
    collaborator BOOLEAN NOT NULL,
    email_verified BOOLEAN NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_associated_user_access FOREIGN KEY (online_access_info_id) REFERENCES shopify_online_access_infos(id)
);
