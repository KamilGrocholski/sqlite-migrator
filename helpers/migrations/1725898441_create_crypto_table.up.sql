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
