-- migration down: add_created_by_columns_in_products
ALTER TABLE products DROP COLUMN created_by;