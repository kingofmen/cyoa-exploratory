-- +goose Up
-- +goose StatementBegin
CREATE TABLE Locations (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content MEDIUMTEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE Locations
-- +goose StatementEnd
