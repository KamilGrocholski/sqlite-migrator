create table if not exists user_role (
    id integer primary key autoincrement,
    name text not null unique,
    created_at datetime default current_timestamp,
    deleted_at datetime null
);
