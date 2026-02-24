-- migration up: add_created_by_columns_in_products
ALTER TABLE products ADD COLUMN created_by UUID NOT NULL;

CREATE INDEX IF NOT EXISTS idx_products_created_by ON products (created_by);
CREATE INDEX IF NOT EXISTS idx_products_created_by_deleted_at ON products (created_by, deleted_at);