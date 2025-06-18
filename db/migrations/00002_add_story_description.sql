-- +goose Up
-- +goose StatementBegin
ALTER TABLE Stories
ADD description MEDIUMTEXT,
ADD start_location BIGINT UNSIGNED;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE Stories
DROP COLUMN description,
DROP COLUMN start_location;
-- +goose StatementEnd
