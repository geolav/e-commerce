-- +goose Up
-- +goose StatementBegin

-- Создаём схему loms если не существует
CREATE SCHEMA IF NOT EXISTS loms;

-- Создаём функцию для автообновления updated_at
CREATE OR REPLACE FUNCTION loms.update_updated_at_column()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $func$
BEGIN
    NEW.updated_at = NOW();
RETURN NEW;
END;
$func$;

-- +goose StatementEnd

-- Добавляем триггер
DROP TRIGGER IF EXISTS update_orders_updated_at ON loms.orders;
CREATE TRIGGER update_orders_updated_at
    BEFORE UPDATE ON loms.orders
    FOR EACH ROW
    EXECUTE FUNCTION loms.update_updated_at_column();

-- Добавляем CHECK constraints
ALTER TABLE loms.order_items DROP CONSTRAINT IF EXISTS order_items_count_check;
ALTER TABLE loms.order_items ADD CONSTRAINT order_items_count_check CHECK (count > 0);

ALTER TABLE loms.products DROP CONSTRAINT IF EXISTS products_price_check;
ALTER TABLE loms.products ADD CONSTRAINT products_price_check CHECK (price >= 0);

ALTER TABLE loms.stocks DROP CONSTRAINT IF EXISTS stocks_count_check;
ALTER TABLE loms.stocks ADD CONSTRAINT stocks_count_check CHECK (count >= 0);

-- Переделываем outbox на JSONB
-- ALTER TABLE loms.outbox ADD COLUMN IF NOT EXISTS payload JSONB;
--
-- -- Переносим данные
-- UPDATE loms.outbox
-- SET payload = jsonb_build_object('order_id', order_id, 'user_id', user_id, 'status', status)
-- WHERE payload IS NULL;
--
-- -- Делаем колонку NOT NULL
-- ALTER TABLE loms.outbox ALTER COLUMN payload SET NOT NULL;
--
-- -- Удаляем старые колонки
-- ALTER TABLE loms.outbox DROP COLUMN IF EXISTS order_id;
-- ALTER TABLE loms.outbox DROP COLUMN IF EXISTS user_id;
-- ALTER TABLE loms.outbox DROP COLUMN IF EXISTS status;

-- +goose Down
-- +goose StatementBegin

-- Восстанавливаем старые колонки
-- ALTER TABLE loms.outbox ADD COLUMN IF NOT EXISTS order_id BIGINT;
-- ALTER TABLE loms.outbox ADD COLUMN IF NOT EXISTS user_id BIGINT;
-- ALTER TABLE loms.outbox ADD COLUMN IF NOT EXISTS status TEXT;
--
-- UPDATE loms.outbox
-- SET
--     order_id = (payload->>'order_id')::BIGINT,
--     user_id = (payload->>'user_id')::BIGINT,
--     status = payload->>'status'
-- WHERE payload IS NOT NULL;
--
-- ALTER TABLE loms.outbox DROP COLUMN IF EXISTS payload;

-- Удаляем CHECK constraints
ALTER TABLE loms.stocks DROP CONSTRAINT IF EXISTS stocks_count_check;
ALTER TABLE loms.products DROP CONSTRAINT IF EXISTS products_price_check;
ALTER TABLE loms.order_items DROP CONSTRAINT IF EXISTS order_items_count_check;

-- Удаляем триггер и функцию
DROP TRIGGER IF EXISTS update_orders_updated_at ON loms.orders;
DROP FUNCTION IF EXISTS loms.update_updated_at_column();

-- +goose StatementEnd