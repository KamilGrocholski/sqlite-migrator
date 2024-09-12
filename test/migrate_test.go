package main

import (
	"database/sql"
	"os"
	"slices"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/KamilGrocholski/sqlite-utils/internal/migrator"
)

func TestSimpleMigrate(t *testing.T) {
	migrationsDir, err := os.MkdirTemp("", "__test_migrations")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(migrationsDir)
	migrationsTable := "test__migration"
	dbName := ":memory:"

	file, err := os.CreateTemp(migrationsDir, "001_create_user_table.sql")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	_, err = file.Write([]byte(
		`-- migrate: up
        create table if not exists user (
            id integer primary key autoincrement
        );

        -- migrate: down
        drop table if exists user;`))
	if err != nil {
		t.Fatal(err)
	}

	file, err = os.CreateTemp(migrationsDir, "002_create_product_table.sql")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	_, err = file.Write([]byte(
		`-- migrate: up
        create table if not exists product (
            id integer primary key autoincrement
        );

        -- migrate: down
        drop table if exists product;`))
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		t.Fatal(err)
	}

	migrator := migrator.Migrator{
		Dir:   migrationsDir,
		Table: migrationsTable,
		DB:    db,
	}
	err = migrator.Migrate()
	if err != nil {
		t.Fatal(err)
	}

	rows, err := db.Query(
		"select name from sqlite_schema where type='table' and name not like 'sqlite_%'")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	tables := []string{}
	for rows.Next() {
		var table string
		err := rows.Scan(&table)
		if err != nil {
			t.Fatal(err)
		}
		tables = append(tables, table)
	}
	if !slices.Contains(tables, migrationsTable) {
		t.Fatal("no migrations table found")
	}
	if !slices.Contains(tables, "user") {
		t.Fatal("no user table found")
	}
	if !slices.Contains(tables, "product") {
		t.Fatal("no product table found")
	}
}
