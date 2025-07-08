-- +goose Up
-- +goose StatementBegin
CREATE TABLE Playthroughs (
    id SERIAL PRIMARY KEY,
    proto BLOB,
    narration MEDIUMTEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE Playthroughs;
-- +goose StatementEnd
