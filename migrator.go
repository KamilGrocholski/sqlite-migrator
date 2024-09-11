package sqliteutils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Migrator struct {
	DB    *sql.DB
	DIR   string
	TABLE string
}

type Migration struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Filename  string    `json:"filename"`
	Up        string    `json:"up"`
	Down      string    `json:"down"`
	CreatedAt time.Time `json:"created_at"`
}

func (m *Migrator) Migrate() error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = m.createDbMigrationTable(tx)
	if err != nil {
		return fmt.Errorf("error creating migration table: %w", err)
	}

	fileMigrations, err := m.readFileMigrations()
	if err != nil {
		return fmt.Errorf("error reading file migrations: %w", err)
	}

	dbMigrations, err := m.getDbMigrations(tx)
	if err != nil {
		return fmt.Errorf("error getting db migrations: %w", err)
	}

	downDbMigrations := slices.Clone(dbMigrations)
	slices.SortFunc(downDbMigrations, func(a, b Migration) int {
		return int(a.ID - b.ID)
	})

	// lastFileMigration := fileMigrations[len(fileMigrations)-1]
	for _, dbMigration := range downDbMigrations {
		has := slices.ContainsFunc(fileMigrations, func(m Migration) bool {
			return m.ID == dbMigration.ID
		})
		if !has {
			break
		}
		_, err := tx.Exec(dbMigration.Down)
		if err != nil {
			return migrationError(err, dbMigration, "error executing down migration")
		}
		err = m.deleteDbMigration(tx, dbMigration.ID)
		if err != nil {
			return migrationError(err, dbMigration, "error deleting down migration")
		}
		for i, m := range dbMigrations {
			if m.ID == dbMigration.ID {
				dbMigrations = append(dbMigrations[:i], dbMigrations[i+1:]...)
			}
		}
	}

	lastDbMigrationID := uint64(0)
	if len(dbMigrations) != 0 {
		lastDbMigrationID = dbMigrations[len(dbMigrations)-1].ID
	}
	for _, fileMigration := range fileMigrations {
		if fileMigration.ID <= lastDbMigrationID {
			continue
		}
		_, err := tx.Exec(fileMigration.Up)
		if err != nil {
			return migrationError(err, fileMigration, "error executing up migration")
		}
		err = m.insertDbMigration(tx, fileMigration)
		if err != nil {
			return migrationError(err, fileMigration, "error inserting up migration")
		}
	}

	return tx.Commit()
}

func (m *Migrator) deleteDbMigration(tx *sql.Tx, id uint64) error {
	_, err := tx.Exec(fmt.Sprintf(`
        delete from %s
        where id = ?
    `, m.TABLE), id)
	return err
}

func (m *Migrator) createDbMigrationTable(tx *sql.Tx) error {
	_, err := tx.Exec(fmt.Sprintf(`
        create table if not exists %s (
            id integer primary key,
            name text not null,
            filename text not null,
            up text not null,
            down text not null,
            created_at datetime default current_timestamp
        );
    `, m.TABLE))
	if err != nil {
		return err
	}
	return nil
}

func (m *Migrator) insertDbMigration(tx *sql.Tx, migration Migration) error {
	_, err := tx.Exec(fmt.Sprintf(`
        insert into %s (id, name, filename, up, down)
        values (?, ?, ?, ?, ?);
    `, m.TABLE,
	),
		migration.ID,
		migration.Name,
		migration.Filename,
		migration.Up,
		migration.Down,
	)
	if err != nil {
		return err
	}
	return nil
}

func (m *Migrator) getDbMigrations(tx *sql.Tx) ([]Migration, error) {
	rows, err := tx.Query(fmt.Sprintf(`
        select * from %s
        order by id;
    `, m.TABLE))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	migrations := []Migration{}
	for rows.Next() {
		var m Migration
		err := rows.Scan(
			&m.ID,
			&m.Name,
			&m.Filename,
			&m.Up,
			&m.Down,
			&m.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}

	return migrations, nil
}

func (m *Migrator) readFileMigrations() ([]Migration, error) {
	entries, err := os.ReadDir(m.DIR)
	if err != nil {
		return nil, err
	}

	migrations := []Migration{}
	for _, entry := range entries {
		if !entry.IsDir() {
			migration, err := m.parseMigration(entry)
			if err != nil {
				return nil, err
			}
			migrations = append(migrations, migration)
		}
	}
	slices.SortFunc(migrations, func(a, b Migration) int {
		return int(a.ID - b.ID)
	})

	return migrations, nil
}

func (m *Migrator) parseMigration(entry fs.DirEntry) (Migration, error) {
	rawID, rest, ok := strings.Cut(entry.Name(), "_")
	if !ok {
		return Migration{}, fmt.Errorf("invalid migration filename: '%s'", entry.Name())
	}
	id, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		return Migration{}, fmt.Errorf("invalid migration filename: %v", err)
	}
	name, _, ok := strings.Cut(rest, ".")
	if !ok {
		return Migration{}, fmt.Errorf("invalid migration filename: '%s'", entry.Name())
	}

	content, err := os.ReadFile(filepath.Join(m.DIR, entry.Name()))
	if err != nil {
		return Migration{}, err
	}
	up, down, ok := strings.Cut(string(content), "-- migrate: down\n")
	if !ok {
		return Migration{}, fmt.Errorf("invalid migration content: '-- migrate: down' not found")
	}
	up, ok = strings.CutPrefix(up, "-- migrate: up\n")
	if !ok {
		return Migration{}, fmt.Errorf("invalid migration content: '-- migrate: up' not found")
	}
	re := regexp.MustCompile(`\s+`)
	up = re.ReplaceAllString(up, " ")
	down = re.ReplaceAllString(down, " ")

	migration := Migration{
		ID:       id,
		Name:     name,
		Filename: entry.Name(),
		Up:       up,
		Down:     down,
	}
	return migration, nil
}

func migrationError(cause error, migration Migration, desc string) error {
	j, err := json.MarshalIndent(migration, "", "   ")
	if err != nil {
		return err
	}
	return fmt.Errorf("%s: %w\nmigration: %s", desc, cause, string(j))
}
