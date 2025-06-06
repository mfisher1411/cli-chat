-- +goose Up
-- +goose StatementBegin
create table "user" (
    id bigserial not null primary key,
    name text not null,
    email text not null,
    role int not null,
    created_at timestamp not null default now(),
    updated_at timestamp
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table "user";
-- +goose StatementEnd
