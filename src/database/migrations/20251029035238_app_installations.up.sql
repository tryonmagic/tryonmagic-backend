CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- App Installations table
CREATE TABLE app_installations(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop VARCHAR(255) NOT NULL UNIQUE,
    access_token VARCHAR(500) NOT NULL,
    scope VARCHAR(500) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    installed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    uninstalled_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Create indexes for app installations
CREATE INDEX idx_app_installations_shop ON app_installations(shop);

CREATE INDEX idx_app_installations_is_active ON app_installations(is_active);

CREATE INDEX idx_app_installations_installed_at ON app_installations(installed_at);