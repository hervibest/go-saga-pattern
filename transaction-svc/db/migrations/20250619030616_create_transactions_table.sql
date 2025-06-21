-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'transaction_status') THEN
        CREATE TYPE transaction_status AS ENUM (
            'PENDING',
			'EXPIRED',
			'SUCCESS',
			'CANCELED',
			'REFUNDING',
			'REFUNDED',
			'FAILED'
        );
    END IF;
END$$;

DO $$
BEGIN 
	IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'internal_status') THEN
		CREATE TYPE internal_status AS ENUM (
			'PENDING',
			'TOKEN_READY',
			'EXPIRED',
			'EXPIRED_CHECKED_INVALID',
			'EXPIRED_CHECKED_VALID',
			'LATE_SETTLEMENT',
			'SETTLED',
			'CANCELED_BY_SYSTEM',
			'CANCELED_BY_USER',
			'FAILED'
		);
	END IF;
END$$;

DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'external_status') THEN
		CREATE TYPE external_status AS ENUM (
			'capture',
			'settlement',
			'pending',
			'deny',
			'cancel',
			'expire',
			'failure'
		);
	END IF;
END$$;



CREATE TABLE IF NOT EXISTS transactions (
	id UUID NOT NULL default uuid_generate_v4(),
	user_id UUID NOT NULL,
	total_price NUMERIC(19,2) NOT NULL CHECK(total_price > 0),
	transaction_status transaction_status NOT NULL DEFAULT 'PENDING',
	internal_status internal_status NOT NULL DEFAULT 'PENDING',
	external_status external_status,
	external_settlement_at TIMESTAMPTZ,
	external_callback_response JSONB,
	snap_token TEXT,
	checkout_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
	payment_at TIMESTAMPTZ,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
	PRIMARY KEY(id)
);


CREATE INDEX idx_transactions_user_id ON transactions (user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transactions;
DROP INDEX IF EXISTS idx_transactions_user_id;
DROP TYPE transaction_status;
DROP TYPE internal_status;
DROP TYPE external_status;
-- +goose StatementEnd
