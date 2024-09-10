package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"

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
	migrationsdir := flag.String("dir", "./migrations", "migrations dir path")
	schema := flag.String("schema", "./schema.sql", "schema path")
	dbconn := flag.String("db", "test.db", "db connection")
	flag.Parse()
	args := flag.Args()

	config := migrator.Config{
		MigrationsDirPath: *migrationsdir,
		SchemaPath:        *schema,
		DBConn:            *dbconn,
	}

	jsonConfig, _ := json.Marshal(config)
	fmt.Println(string(jsonConfig))

	db, err := sql.Open("sqlite3", config.DBConn)
	if err != nil {
		return err
	}
	defer db.Close()

	migrator := migrator.New(config, db)

	fmt.Println(args)

	// TODO
	switch len(args) {
	case 2:
		action := args[0]
		switch action {
		case "create":
			desc := args[1]
			err = migrator.CreateMigration(desc)
			if err != nil {
				return err
			}
			break
		case "migrate":
			direction := args[1]
			switch direction {
			case "up":
				err = migrator.MigrateUp()
				if err != nil {
					return err
				}
				break
			}
			break
		}
		break
	}

	return nil
}
