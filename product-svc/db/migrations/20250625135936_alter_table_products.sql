-- +goose Up
-- +goose StatementBegin
-- Hapus constraint unik lama
ALTER TABLE products DROP CONSTRAINT IF EXISTS products_name_key;
ALTER TABLE products DROP CONSTRAINT IF EXISTS products_slug_key;

-- Buat partial unique index hanya untuk produk yang belum dihapus
CREATE UNIQUE INDEX unique_product_name_not_deleted 
ON products(name)
WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX unique_product_slug_not_deleted 
ON products(slug)
WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Hapus partial unique index
DROP INDEX IF EXISTS unique_product_name_not_deleted;
DROP INDEX IF EXISTS unique_product_slug_not_deleted;

-- Kembalikan constraint unik pada kolom name
ALTER TABLE products ADD CONSTRAINT products_name_key UNIQUE(name);
ALTER TABLE products ADD CONSTRAINT products_slug_key UNIQUE(slug);
-- +goose StatementEnd
