package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/KamilGrocholski/sqlite-utils/internal/migrator"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	migrationsDir := flag.String("dir", "migrations", "migrations directory")
	migrationsTable := flag.String("table", "__migration", "migrations table name")
	dbName := flag.String("db", ":memory:", "sqlite filepath")
	flag.Parse()

	db, err := sql.Open("sqlite3", *dbName)
	if err != nil {
		return err
	}

	migrator := migrator.Migrator{
		Dir:   *migrationsDir,
		Table: *migrationsTable,
		DB:    db,
	}
	err = migrator.Migrate()
	if err != nil {
		return err
	}

	return nil
}
