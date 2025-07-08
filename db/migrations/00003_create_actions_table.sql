-- +goose Up
-- +goose StatementBegin
CREATE TABLE Actions (
    id CHAR(36) PRIMARY KEY,
    proto BLOB
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE Actions;
-- +goose StatementEnd
