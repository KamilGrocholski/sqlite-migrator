create table if not exists user (
    id integer primary key autoincrement,
    email text not null unique,
    password text not null,
    created_at datetime default current_timestamp,
    updated_at datetime default current_timestamp,
    deleted_at datetime null
);

create index if not exists idx_user_id on user(id);
create index if not exists idx_user_email on user(email);
create index if not exists idx_user_deleted_at on user(deleted_at);

create trigger if not exists tr_user_update
    after update
    on user
begin
    update user
    set updated_at = current_timestamp
    where id = old.id;
end;
