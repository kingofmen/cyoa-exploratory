-- +goose Up
-- +goose StatementBegin
CREATE TABLE StoryLocations (
    story_id BIGINT UNSIGNED NOT NULL,
    location_id CHAR(36) NOT NULL,

    PRIMARY KEY (story_id, location_id),
    FOREIGN KEY (story_id) REFERENCES Stories(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (location_id) REFERENCES Locations(id) ON DELETE CASCADE ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE StoryLocations;
-- +goose StatementEnd
