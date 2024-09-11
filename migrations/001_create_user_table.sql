-- migrate: up
create table if not exists user_role (
    id integer primary key autoincrement,
    name text not null unique
);

create table if not exists user (
    id integer primary key autoincrement
);

-- migrate: down
drop table if exists user;
