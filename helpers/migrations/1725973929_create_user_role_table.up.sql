create table if not exists user_role (
    id integer primary key autoincrement,
    name text not null unique,
    created_at datetime default current_timestamp,
    deleted_at datetime null
);

create index if not exists idx_user_role_id on user_role(id);
create index if not exists idx_user_role_name on user_role(name);
create index if not exists idx_user_role_deleted_at on user_role(deleted_at);
