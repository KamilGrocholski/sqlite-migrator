-- migrate: up
create table if not exists product_order_status (
    id integer primary key autoincrement,
    name text not null unique
);

create table if not exists product_order (
    id integer primary key autoincrement,
    product_order_status_id integer references product_order_status(id)
);

-- migrate: down
drop table if exists product_order_status;
drop table if exists product_order;
