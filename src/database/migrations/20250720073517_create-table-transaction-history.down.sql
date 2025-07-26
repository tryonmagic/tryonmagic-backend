-- Drop the transaction_history table
DROP TABLE IF EXISTS transaction_history;

-- Drop the ENUM types
DROP TYPE IF EXISTS transaction_type_enum;

DROP TYPE IF EXISTS transaction_status_enum;

DROP TYPE IF EXISTS payment_method_enum;

DROP TYPE IF EXISTS payment_status_enum;

DROP TYPE IF EXISTS reference_type_enum;