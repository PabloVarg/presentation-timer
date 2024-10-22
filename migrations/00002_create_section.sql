-- +goose Up
-- +goose StatementBegin
CREATE TABLE section (
    id BIGSERIAL PRIMARY KEY,
    presentation BIGINT REFERENCES presentation(id) ON DELETE CASCADE,

    name TEXT NOT NULL,
    duration INTERVAL NOT NULL,
    position SMALLINT NOT NULL DEFAULT 1::SMALLINT
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE section;
-- +goose StatementEnd
