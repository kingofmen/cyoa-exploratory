-- +goose Up
-- +goose StatementBegin
CREATE TABLE Locations (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    proto BLOB
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE Locations
-- +goose StatementEnd
