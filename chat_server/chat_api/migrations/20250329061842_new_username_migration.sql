-- +goose Up
create table if not exists usernames (
    id serial primary key,
    usernames text not null
);

-- +goose Down
drop table if exists usernames;

