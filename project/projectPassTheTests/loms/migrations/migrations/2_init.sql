-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS loms;

CREATE TYPE loms.order_status AS ENUM (
    'new',
    'awaiting_payment',
    'failed',
    'paid',
    'cancelled'
);

CREATE TABLE IF NOT EXISTS loms.products (
    sku     SERIAL PRIMARY KEY,
    name    TEXT NOT NULL,
    price   INTEGER NOT NULL CHECK (price >= 0)
);

CREATE TABLE IF NOT EXISTS loms.stocks (
    sku     INTEGER PRIMARY KEY REFERENCES loms.products(sku),
    count   BIGINT NOT NULL DEFAULT 0
);

-- CREATE TABLE IF NOT EXISTS loms.orders (
--     id          BIGSERIAL PRIMARY KEY,
--     user_id     BIGINT NOT NULL,
--     status      loms.order_status NOT NULL DEFAULT 'new',
--     created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--     updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
-- );

CREATE SEQUENCE IF NOT EXISTS loms.order_id_seq;

CREATE TABLE IF NOT EXISTS loms.orders (
    id          BIGINT PRIMARY KEY DEFAULT nextval('loms.order_id_seq'),
    user_id     BIGINT NOT NULL,
    status      loms.order_status NOT NULL DEFAULT 'new',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );


CREATE TABLE IF NOT EXISTS loms.order_items (
    order_id    BIGINT NOT NULL REFERENCES loms.orders(id),
    sku         INTEGER NOT NULL,
    count       INTEGER NOT NULL CHECK (count > 0),
    PRIMARY KEY (order_id, sku)
);

-- CREATE TABLE IF NOT EXISTS loms.outbox (
--     id              BIGSERIAL PRIMARY KEY,
--     order_id        BIGINT NOT NULL,
--     user_id         BIGINT NOT NULL,
--     status          TEXT NOT NULL,
--     created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--     attempts        INT NOT NULL DEFAULT 0,
--     in_progress_at  TIMESTAMPTZ
-- );

CREATE TABLE IF NOT EXISTS loms.outbox (
    id              BIGSERIAL PRIMARY KEY,
    payload         JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    attempts        INT NOT NULL DEFAULT 0,
    in_progress_at  TIMESTAMPTZ
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS loms.outbox;
DROP TABLE IF EXISTS loms.order_items;
DROP TABLE IF EXISTS loms.orders;
DROP TABLE IF EXISTS loms.stocks;
DROP TABLE IF EXISTS loms.products;
DROP SEQUENCE IF EXISTS loms.order_id_seq;
DROP TYPE IF EXISTS loms.order_status;
DROP SCHEMA IF EXISTS loms;
-- +goose StatementEnd