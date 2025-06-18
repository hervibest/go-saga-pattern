-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS products (
	id UUID NOT NULL UNIQUE default uuid_generate_v4(),
    user_id UUID NOT NULL,
	name VARCHAR(255) NOT NULL UNIQUE,
	slug VARCHAR(255) NOT NULL UNIQUE,
	description TEXT,
	price NUMERIC(19,2) NOT NULL CHECK(price > 0),
	quantity INTEGER NOT NULL DEFAULT 0,
	created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
	deleted_at TIMESTAMPTZ,
	PRIMARY KEY(id)
);

CREATE INDEX idx_products_user_id_deleted_at ON products (user_id, deleted_at);
CREATE INDEX idx_products_id_deleted_at ON products (id, deleted_at);
CREATE INDEX idx_products_slug_deleted_at ON products (slug, deleted_at);
CREATE INDEX idx_products_id_user_id_deleted_at ON products (id, user_id, deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS products;
DROP INDEX IF EXISTS idx_products_user_id_deleted_at;
DROP INDEX IF EXISTS idx_products_id_deleted_at;
DROP INDEX IF EXISTS idx_products_slug;
DROP INDEX IF EXISTS idx_products_id_user_id_deleted_at;

-- +goose StatementEnd
