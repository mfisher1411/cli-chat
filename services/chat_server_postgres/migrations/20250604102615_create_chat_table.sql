-- +goose Up
-- +goose StatementBegin
create table chat (
    id bigserial not null primary key
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table chat;
-- +goose StatementEnd
