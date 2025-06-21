-- +goose Up
-- +goose StatementBegin
CREATE TABLE transaction_details (
	id UUID NOT NULL default uuid_generate_v4(),
	transaction_id UUID NOT NULL,
	product_id UUID NOT NULL,
	quantity INTEGER NOT NULL CHECK(quantity > 0),
	price NUMERIC(19,2) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
	PRIMARY KEY(id)
);

CREATE INDEX idx_transaction_details_transaction_id ON transaction_details (transaction_id);
CREATE INDEX idx_transaction_details_product_id ON transaction_details (product_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transaction_details;
DROP INDEX IF EXISTS idx_transaction_details_transaction_id;
DROP INDEX IF EXISTS idx_transaction_details_product_id;
-- +goose StatementEnd
