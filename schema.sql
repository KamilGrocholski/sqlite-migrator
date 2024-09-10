create table if not exists user (
    id integer primary key autoincrement,
    email text not null unique,
    password text not null,
    created_at datetime default current_timestamp
);

create table if not exists user_role (
    id integer primary key autoincrement,
    name text not null unique,
    created_at datetime default current_timestamp,
    deleted_at datetime null
);

alter table user
add role_id integer references user_role(id);

