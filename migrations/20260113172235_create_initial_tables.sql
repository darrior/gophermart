-- +goose Up
-- +goose StatementBegin
CREATE TABLE users(
    id uuid NOT NULL PRIMARY KEY,
    login varchar(32),
    password_hash text,
    current_balance float,
    withdrawan_balance float
);
CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
CREATE TABLE orders(
    number varchar(64) PRIMARY KEY,
    accrual float,
    user_uuid uuid,
    status order_status,
    uploaded_at timestamp
);
CREATE TABLE withdrawals(
    id SERIAL PRIMARY KEY,
    user_uuid uuid,
    order_number varchar(64),
    sum float,
    processed_at timestamp
);
CREATE UNIQUE INDEX unique_numbers ON orders (number);
CREATE UNIQUE INDEX unique_logins ON users (login);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS withdrawals;
DROP TABLE IF EXISTS orders;
DROP TYPE IF EXISTS order_status;
DROP TABLE IF EXISTS users;
DROP INDEX IF EXISTS unique_logins;
DROP INDEX IF EXISTS unique_numbers;
-- +goose StatementEnd
