-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS cart;

CREATE TABLE IF NOT EXISTS cart.items (
    user_id     BIGINT NOT NULL,
    sku         INTEGER NOT NULL,
    count       INTEGER NOT NULL CHECK (count > 0),
    PRIMARY KEY (user_id, sku)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cart.items;
DROP SCHEMA IF EXISTS cart;
-- +goose StatementEnd