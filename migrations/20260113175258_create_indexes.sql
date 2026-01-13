-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX unique_numbers ON orders (number);
CREATE UNIQUE INDEX unique_logins ON users (logins);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS unique_logins;
DROP INDEX IF EXISTS unique_numbers;
-- +goose StatementEnd
