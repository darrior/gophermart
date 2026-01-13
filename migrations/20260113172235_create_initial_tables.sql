-- +goose Up
-- +goose StatementBegin
CREATE TABLE users(
    id uuid NOT NULL,
    login text,
    password_hash text,
    current_balance float,
    withdrawan_balance float,
    PRIMARY KEY(id)
);
CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
CREATE TABLE orders(
    number text,
    accural float,
    user_uuid uuid,
    status order_status,
    uploaded_at timestamp,
    PRIMARY KEY(number)
);
CREATE TABLE withdrawals(
    id SERIAL PRIMARY KEY ,
    user_uuid uuid,
    order text,
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
