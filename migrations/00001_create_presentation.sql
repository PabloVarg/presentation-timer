-- +goose Up
-- +goose StatementBegin
CREATE TABLE presentation (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE presentation;
-- +goose StatementEnd
