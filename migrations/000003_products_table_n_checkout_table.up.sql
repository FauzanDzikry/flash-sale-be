-- migration up: products_table_n_checkout_table
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    category VARCHAR(255) NOT NULL,
    stock INT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    discount DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS checkouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    product_id UUID NOT NULL,
    quantity INT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    discount DECIMAL(10, 2) NOT NULL,
    total_price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_checkouts_user_id ON checkouts (user_id);
CREATE INDEX IF NOT EXISTS idx_checkouts_product_id ON checkouts (product_id);
CREATE INDEX IF NOT EXISTS idx_checkouts_user_id_product_id ON checkouts (user_id, product_id);
CREATE INDEX IF NOT EXISTS idx_checkouts_user_id_product_id_deleted_at ON checkouts (user_id, product_id, deleted_at);