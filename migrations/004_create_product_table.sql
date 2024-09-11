-- migrate: up
create table if not exists product (
    id integer primary key autoincrement,
    name text not null,
    description text,
    created_at datetime default current_timestamp,
    updated_at datetime default current_timestamp,
    deleted_at datetime null
);

create index if not exists idx_product_id on product(id);
create index if not exists idx_product_name on product(name);
create index if not exists idx_product_deleted on product(deleted_at);

create trigger if not exists tr_product_update 
    after update
    on product
begin
    update product
    set updated_at = current_timestamp
    where id = old.id;
end;

-- migrate: down
drop table if exists product;
