-- +goose Up
-- +goose StatementBegin
CREATE TABLE Stories (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    proto BLOB
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE Stories
-- +goose StatementEnd
