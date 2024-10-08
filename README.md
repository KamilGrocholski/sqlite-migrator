# sqlite-migrator

## Build

```bash
go build -o migrate cmd/migrate/main.go
```

## Create '001_create_user_table.sql' file inside 'migrations_dir'

```bash
-- migrate: up
create table if not exists user (
    id integer primary key autoincrement
);

-- migrate: down
drop table if exists user;
```

## Run

```bash
./migrate -db sqlite_file -dir migrations_dir -table __migration
```

## Check the results

### Connect to the db

```bash
sqlite3 sqlite_file
```

### Show tables

```bash
.tables
```

### Show migration rows in '\_\_migration' table

```bash
select * from __migration;
```
