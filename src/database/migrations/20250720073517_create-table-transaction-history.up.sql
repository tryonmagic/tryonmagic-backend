-- Create ENUM types for transaction fields
DO $$ BEGIN CREATE TYPE transaction_type_enum AS ENUM (
    'TASK_EXECUTION',
    'SUBSCRIPTION_PURCHASE',
    'POINT_PURCHASE',
    'REFUND',
    'BONUS',
    'PLAN_UPGRADE',
    'PLAN_DOWNGRADE',
    'MONTHLY_RESET',
    'REFERRAL_BONUS',
    'PROMOTIONAL_CREDIT',
    'COMPENSATION',
    'ADJUSTMENT'
);

EXCEPTION
WHEN duplicate_object THEN null;

END $$;

DO $$ BEGIN CREATE TYPE transaction_status_enum AS ENUM (
    'PENDING',
    'COMPLETED',
    'FAILED',
    'CANCELLED'
);

EXCEPTION
WHEN duplicate_object THEN null;

END $$;

DO $$ BEGIN CREATE TYPE payment_method_enum AS ENUM (
    'CREDIT_CARD',
    'PAYPAL',
    'STRIPE',
    'BANK_TRANSFER',
    'FREE',
    'CRYPTO',
    'SHOPIFY',
    'PARTNER'

);

EXCEPTION
WHEN duplicate_object THEN null;

END $$;

DO $$ BEGIN CREATE TYPE payment_status_enum AS ENUM (
    'PENDING',
    'COMPLETED',
    'FAILED',
    'REFUNDED',
    'CANCELLED'
);

EXCEPTION
WHEN duplicate_object THEN null;

END $$;

DO $$ BEGIN CREATE TYPE reference_type_enum AS ENUM (
    'PAYMENT',
    'ORDER',
    'TASK',
    'SUBSCRIPTION',
    'REFUND',
    'BONUS'
);

EXCEPTION
WHEN duplicate_object THEN null;

END $$;

CREATE TABLE IF NOT EXISTS transaction_histories (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    task_id UUID NULL,
    transaction_type transaction_type_enum NOT NULL,
    transaction_status transaction_status_enum NOT NULL DEFAULT 'COMPLETED',
    debit_amount INT NOT NULL DEFAULT 0,
    credit_amount INT NOT NULL DEFAULT 0,
    net_amount INT NOT NULL,
    balance_before INT NOT NULL,
    balance_after INT NOT NULL,
    description VARCHAR(500) NOT NULL,
    reference_id VARCHAR(255) NULL,
    reference_type reference_type_enum NULL,
    payment_method payment_method_enum NULL,
    payment_status payment_status_enum NULL,
    currency VARCHAR(10) DEFAULT 'USD',
    amount_paid DECIMAL(10, 2) DEFAULT 0.00,
    exchange_rate DECIMAL(10, 4) DEFAULT 1.0000,
    metadata TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP NULL,
    expires_at TIMESTAMP NULL,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (task_id) REFERENCES tasks(id)
);