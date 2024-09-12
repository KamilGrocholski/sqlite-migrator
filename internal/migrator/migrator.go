package migrator

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Migrator struct {
	Table string
	Dir   string
	DB    *sql.DB
}

type Migration struct {
	ID       uint64 `json:"id"`
	Filename string `json:"filename"`
	Up       string `json:"up"`
	Down     string `json:"down"`
}

func (m *Migration) Pretty() (string, error) {
	pretty, err := json.MarshalIndent(m, "", "    ")
	return string(pretty), err
}

func migrationError(cause error, desc string, migration *Migration) error {
	pretty, err := migration.Pretty()
	if err != nil {
		return fmt.Errorf("desc: %s\ncause: %w\nmigration:%w\n", desc, cause, err)
	}
	return fmt.Errorf("desc: %s\ncause: %w\nmigration:\n%s\n", desc, cause, pretty)
}

type MigrationsMap map[uint64]*Migration

func (m *Migrator) Migrate() error {
	fileMigrations, fileMigrationsMap, err := m.readFileMigrations()
	if err != nil {
		return err
	}

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := m.dbUpsertMigrationTable(tx); err != nil {
		return fmt.Errorf("failed_db_migration_table_upsert: %w", err)
	}

	dbMigrations, err := m.dbGetMigrations(tx)
	if err != nil {
		return fmt.Errorf("failed_db_get_migrations: %w", err)
	}
	dbMigrationsUntouchedMap := make(MigrationsMap)
	for i := len(dbMigrations) - 1; i >= 0; i-- {
		dbMigration := dbMigrations[i]
		if _, ok := fileMigrationsMap[dbMigration.ID]; ok {
			dbMigrationsUntouchedMap[dbMigration.ID] = dbMigration
			continue
		}
		_, err := tx.Exec(dbMigration.Down)
		if err != nil {
			return migrationError(err, "failed_down_db_migration_execution", dbMigration)
		}
		err = m.dbDeleteMigration(tx, dbMigration.ID)
		if err != nil {
			return migrationError(err, "failed_db_db_migration_deleting", dbMigration)
		}
	}

	for _, fileMigration := range fileMigrations {
		if _, ok := dbMigrationsUntouchedMap[fileMigration.ID]; ok {
			continue
		}
		_, err := tx.Exec(fileMigration.Up)
		if err != nil {
			return migrationError(err, "failed_up_file_migration_execution", fileMigration)
		}
		err = m.dbInsertMigration(tx, fileMigration)
		if err != nil {
			return migrationError(err, "failed_up_file_migration_insert", fileMigration)
		}
	}

	return tx.Commit()
}

func (m *Migrator) readFileMigrations() ([]*Migration, MigrationsMap, error) {
	entries, err := os.ReadDir(m.Dir)
	if err != nil {
		return nil, nil, err
	}

	migrations := make([]*Migration, 0, len(entries))
	migrationsMap := make(MigrationsMap)
	for _, entry := range entries {
		migration, err := m.parseFileMigrationEntry(entry)
		if err != nil {
			return nil, nil, err
		}
		migrations = append(migrations, migration)
		migrationsMap[migration.ID] = migration
	}

	return migrations, migrationsMap, nil
}

func (m *Migrator) parseFileMigrationEntry(entry fs.DirEntry) (*Migration, error) {
	rawID, rest, ok := strings.Cut(entry.Name(), "_")
	if !ok {
		return nil, fmt.Errorf("invalid_migration_file_name_format: '%s'", entry.Name())
	}
	id, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid_migration_file_name_format: '%s'", entry.Name())
	}
	_, _, ok = strings.Cut(rest, ".")
	if !ok {
		return nil, fmt.Errorf("invalid_migration_file_name_format: '%s'", entry.Name())
	}

	content, err := os.ReadFile(filepath.Join(m.Dir, entry.Name()))
	if err != nil {
		return nil, fmt.Errorf("failed_reading_migration_file: %w", err)
	}
	up, down, ok := strings.Cut(string(content), "-- migrate: down")
	if !ok {
		return nil, fmt.Errorf("invalid_migration_file_content_format: expected '-- migrate: down'")
	}
	down = strings.TrimSpace(down)
	up, ok = strings.CutPrefix(up, "-- migrate: up")
	if !ok {
		return nil, fmt.Errorf("invalid_migration_file_content_format: expected '-- migrate: up'")
	}
	up = strings.TrimSpace(up)

	migration := &Migration{
		ID:       id,
		Filename: entry.Name(),
		Up:       up,
		Down:     down,
	}

	return migration, nil
}

func (m *Migrator) dbUpsertMigrationTable(tx *sql.Tx) error {
	_, err := tx.Exec(fmt.Sprintf(`
        create table if not exists %s (
            id integer primary key,
            filename text not null,
            up text not null,
            down text not null
        );
    `, m.Table))
	return err
}

func (m *Migrator) dbGetMigrations(tx *sql.Tx) ([]*Migration, error) {
	var migrations []*Migration
	rows, err := tx.Query(fmt.Sprintf(`
        select id, filename, up, down from %s order by id ASC;
    `, m.Table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var m Migration
		err := rows.Scan(
			&m.ID,
			&m.Filename,
			&m.Up,
			&m.Down,
		)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, &m)
	}
	return migrations, nil
}

func (m *Migrator) dbDeleteMigration(tx *sql.Tx, id uint64) error {
	_, err := tx.Exec(fmt.Sprintf(`
        delete from %s where id = ?;
    `, m.Table), id)
	return err
}

func (m *Migrator) dbInsertMigration(tx *sql.Tx, migration *Migration) error {
	_, err := tx.Exec(fmt.Sprintf(`
        insert into %s (id, filename, up, down)
        values (?, ?, ?, ?);
    `, m.Table),
		migration.ID,
		migration.Filename,
		migration.Up,
		migration.Down,
	)
	return err
}
