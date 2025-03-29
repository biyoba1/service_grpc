-- +goose Up
create table if not exists chat (
    id serial primary key,
    fromm text not null,
    msg text not null,
    time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
drop table if exists chat;

