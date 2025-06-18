-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS product_transactions (
	transaction_id UUID NOT NULL,
	product_id UUID NOT NULL,
	-- RESERVED, CANCELED, COMMITED, SETTLED
	status VARCHAR(100) NOT NULL DEFAULT 'RESERVED',
	quantity INTEGER NOT NULL,
	total_price NUMERIC(19,2) NOT NULL CHECK(total_price > 0),
	reserved_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
	canceled_at TIMESTAMPTZ,
	comitted_at TIMESTAMPTZ,
	expired_at TIMESTAMPTZ,
	settled_at TIMESTAMPTZ,
	PRIMARY KEY(transaction_id, product_id)
);COMMENT ON COLUMN product_transactions.status IS 'RESERVED, CANCELED, COMMITED, EXPIRED, SETTLED';

CREATE INDEX idx_product_transactions_status ON product_transactions (status, product_id, transaction_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS product_transactions;
DROP INDEX IF EXISTS idx_product_transactions_status;
-- +goose StatementEnd
