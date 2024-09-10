create table if not exists user (
    id integer primary key autoincrement,
    email text not null unique,
    password text not null,
    created_at datetime default current_timestamp
);
