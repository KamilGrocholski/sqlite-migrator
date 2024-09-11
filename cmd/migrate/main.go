package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	sqliteutils "github.com/KamilGrocholski/sqlite-utils"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	migrationsTable := flag.String("table", "__migration", "migration table name")
	migrationsDir := flag.String("migrations", "migrations", "migrations directory")
	sqliteFilePath := flag.String("db", ":memory:", "sqlite file path")
	flag.Parse()

	db, err := sql.Open("sqlite3", *sqliteFilePath)
	if err != nil {
		return err
	}

	migrator := sqliteutils.Migrator{
		DB:    db,
		DIR:   *migrationsDir,
		TABLE: *migrationsTable,
	}

	err = migrator.Migrate()
	if err != nil {
		return err
	}

	return nil
}
