-- +goose Up
-- +goose StatementBegin
CREATE TABLE StoryActions (
    story_id BIGINT UNSIGNED NOT NULL,
    action_id CHAR(36) NOT NULL,
    PRIMARY KEY (story_id, action_id),
    FOREIGN KEY (story_id) REFERENCES Stories(id) ON DELETE CASCADE,
    FOREIGN KEY (action_id) REFERENCES Actions(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE StoryActions;
-- +goose StatementEnd
