-- +goose Up
CREATE SCHEMA IF NOT EXISTS cart;

CREATE TABLE IF NOT EXISTS cart.items (
    user_id BIGINT NOT NULL,
    sku BIGINT NOT NULL,
    count INTEGER NOT NULL,
    PRIMARY KEY (user_id, sku)
);

-- +goose Down
DROP TABLE IF EXISTS cart.items;
DROP SCHEMA IF EXISTS cart;
