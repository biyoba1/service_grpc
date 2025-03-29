-- +goose Up
create table if not exists auth (
    id serial primary key,
    name text,
    email text,
    password text,
    role text,
    created_at timestamp not null default now(),
    updated_at TIMESTAMP
);

-- +goose Down
drop table if exists auth;
