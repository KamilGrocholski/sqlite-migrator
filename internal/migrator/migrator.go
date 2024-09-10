package migrator

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Migrator interface {
	UserVersion() (uint, error)
	CountMigrations() (uint, error)
	CreateMigration(desc string) error
	MigrateUp() error
	// MigrateDown(migration string) error
}

const (
	MigrationUpExt   = "up.sql"
	MigrationDownExt = "down.sql"
)

func composeMigrationFilename(timestamp int64, desc string, ext string) string {
	return fmt.Sprintf("%d_%s.%s", timestamp, desc, ext)
}

type migrator struct {
	mu     *sync.RWMutex
	db     *sql.DB
	config Config
}

type Config struct {
	MigrationsDirPath string `json:"migrations_dir_path"`
	SchemaPath        string `json:"schema_path"`
	DBConn            string `json:"db_conn"`
}

func New(
	config Config,
	db *sql.DB,
) Migrator {
	return &migrator{
		mu:     &sync.RWMutex{},
		config: config,
		db:     db,
	}
}

func (m *migrator) MigrateUp() error {
	version, err := m.UserVersion()
	if err != nil {
		return err
	}

	upMigrationsEntries, _, migrationsCount, err := m.getSortedMigrationEntries()
	if migrationsCount == version {
		return nil
	}

	var schema bytes.Buffer

	for _, entry := range upMigrationsEntries {
		bytes, err := os.ReadFile(filepath.Join(m.config.MigrationsDirPath, entry.Name()))
		if err != nil {
			return err
		}
		_, err = schema.Write(bytes)
		if err != nil {
			return err
		}
		err = schema.WriteByte('\n')
		if err != nil {
			return err
		}
	}

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(schema.String())
	if err != nil {
		return err
	}

	_, err = tx.Exec(fmt.Sprintf("pragma user_version=%d", migrationsCount))
	if err != nil {
		return err
	}

	err = os.WriteFile(m.config.SchemaPath, schema.Bytes(), 0664)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (m *migrator) UserVersion() (uint, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	row := m.db.QueryRow("pragma user_version")
	if row.Err() != nil {
		return 0, row.Err()
	}
	var version uint
	err := row.Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

func (m *migrator) CountMigrations() (uint, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, _, count, err := m.getSortedMigrationEntries()
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (m *migrator) CreateMigration(desc string) error {
	now := time.Now().UnixMilli()
	up := composeMigrationFilename(now, desc, MigrationUpExt)
	down := composeMigrationFilename(now, desc, MigrationDownExt)
	upFile, err := os.Create(filepath.Join(m.config.MigrationsDirPath, up))
	if err != nil {
		return err
	}
	_, err = os.Create(filepath.Join(m.config.MigrationsDirPath, down))
	if err != nil {
		os.Remove(upFile.Name())
		return err
	}
	return nil
}

func (m *migrator) getSortedMigrationEntries() ([]fs.DirEntry, []fs.DirEntry, uint, error) {
	entries, err := os.ReadDir(m.config.MigrationsDirPath)
	if err != nil {
		return nil, nil, 0, err
	}

	upEntries := []fs.DirEntry{}
	downEntries := []fs.DirEntry{}
	for _, entry := range entries {
		if !entry.IsDir() {
			if strings.HasSuffix(entry.Name(), MigrationDownExt) {
				downEntries = append(downEntries, entry)
			} else if strings.HasSuffix(entry.Name(), MigrationUpExt) {
				upEntries = append(upEntries, entry)
			}
		}
	}

	slices.SortFunc(upEntries, func(a, b fs.DirEntry) int {
		return strings.Compare(a.Name(), b.Name())
	})

	slices.SortFunc(downEntries, func(a, b fs.DirEntry) int {
		return strings.Compare(b.Name(), a.Name())
	})

	count := uint(len(upEntries))

	return upEntries, downEntries, count, nil
}
