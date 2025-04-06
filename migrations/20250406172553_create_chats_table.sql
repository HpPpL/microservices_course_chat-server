-- +goose Up
-- +goose StatementBegin
CREATE TABLE chats (
    id serial PRIMARY KEY,
    users text[]
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE chats
-- +goose StatementEnd
