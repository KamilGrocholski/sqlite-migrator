-- migrate: up
create table if not exists crypto (
    id integer primary key autoincrement
);

-- migrate: down
drop table if exists crypto;
