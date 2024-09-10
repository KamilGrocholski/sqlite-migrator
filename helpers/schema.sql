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
create table if not exists crypto (
    id integer primary key autoincrement,
    symbol text not null unique,
    name text not null,
    created_at datetime default current_timestamp,
    updated_at datetime default current_timestamp,
    deleted_at datetime null
);

create index if not exists idx_crypto_id on crypto(id);
create index if not exists idx_crypto_symbol on crypto(symbol);
create index if not exists idx_crypto_deleted_at on crypto(deleted_at);

create trigger if not exists tr_crypto_update
    after update
    on crypto
begin
    update crypto
    set updated_at = current_timestamp
    where id = old.id;
end;
create table if not exists user_role (
    id integer primary key autoincrement,
    name text not null unique,
    created_at datetime default current_timestamp,
    deleted_at datetime null
);

create index if not exists idx_user_role_id on user_role(id);
create index if not exists idx_user_role_name on user_role(name);
create index if not exists idx_user_role_deleted_at on user_role(deleted_at);
