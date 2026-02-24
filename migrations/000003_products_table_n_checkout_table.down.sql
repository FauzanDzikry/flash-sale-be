-- migration down: products_table_n_checkout_table
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS checkouts;
