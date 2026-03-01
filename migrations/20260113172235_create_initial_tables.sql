-- +goose Up
-- +goose StatementBegin
CREATE TABLE users(
    id uuid NOT NULL PRIMARY KEY,
    login text,
    password_hash text,
    current_balance float,
    withdrawan_balance float
);
CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
CREATE TABLE orders(
    number text PRIMARY KEY,
    accrual float,
    user_uuid uuid,
    status order_status,
    uploaded_at timestamp
);
CREATE TABLE withdrawals(
    id SERIAL PRIMARY KEY,
    user_uuid uuid,
    order_number text,
    sum float,
    processed_at timestamp
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS withdrawals;
DROP TABLE IF EXISTS orders;
DROP TYPE IF EXISTS order_status;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
