package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectDB opens a GORM Postgres connection using cfg.DSN(). Foreign-key
// constraint creation during migration is disabled at the GORM level because
// all schema changes are owned by SQL migrations, not by AutoMigrate.
func ConnectDB(cfg *Config) (*gorm.DB, error) {
	gormLogger := logger.Default.LogMode(logger.Warn)
	if cfg.AppEnv == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("acquire sql.DB: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return db, nil
}

// AssertMigrationsUpToDate verifies that every *.up.sql file under
// cfg.MigrationsDir has been applied (i.e. its numeric version <= the version
// recorded in schema_migrations) and that the schema_migrations row is not
// marked dirty. It NEVER applies migrations — that is reserved for
// `make migrate-up`. The function fails loud if the DB is behind, dirty, or
// missing the schema_migrations table after at least one migration file
// exists on disk.
func AssertMigrationsUpToDate(db *gorm.DB, migrationsDir string) error {
	latestOnDisk, err := latestMigrationVersionOnDisk(migrationsDir)
	if err != nil {
		return fmt.Errorf("scan migrations dir %q: %w", migrationsDir, err)
	}
	if latestOnDisk == 0 {
		// No migrations exist yet — nothing to assert. This only happens
		// before the first migration file is added.
		return nil
	}

	var hasTable bool
	if err := db.Raw(
		`SELECT EXISTS (
            SELECT 1 FROM information_schema.tables
            WHERE table_schema = current_schema() AND table_name = 'schema_migrations'
        )`,
	).Scan(&hasTable).Error; err != nil {
		return fmt.Errorf("check schema_migrations table: %w", err)
	}
	if !hasTable {
		return fmt.Errorf(
			"schema_migrations table not found but %d migration file(s) exist on disk; run `make migrate-up` before starting the server",
			latestOnDisk,
		)
	}

	type row struct {
		Version int64
		Dirty   bool
	}
	var r row
	err = db.Raw(`SELECT version, dirty FROM schema_migrations LIMIT 1`).Scan(&r).Error
	if err != nil {
		return fmt.Errorf("read schema_migrations: %w", err)
	}
	if r.Dirty {
		return fmt.Errorf(
			"schema_migrations is dirty at version %d; fix manually with `make migrate-force version=<N>` then `make migrate-up`",
			r.Version,
		)
	}
	if r.Version < latestOnDisk {
		return fmt.Errorf(
			"database is behind: applied version=%d, latest on disk=%d; run `make migrate-up`",
			r.Version, latestOnDisk,
		)
	}
	return nil
}

// latestMigrationVersionOnDisk returns the highest NNNNNN sequence number
// among *.up.sql files in migrationsDir. Returns 0 if the directory is empty
// or missing.
func latestMigrationVersionOnDisk(migrationsDir string) (int64, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	versions := make([]int64, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		// Filename pattern: NNNNNN_anything.up.sql
		base := filepath.Base(name)
		underscore := strings.IndexByte(base, '_')
		if underscore <= 0 {
			continue
		}
		seqStr := base[:underscore]
		seq, err := strconv.ParseInt(seqStr, 10, 64)
		if err != nil {
			continue
		}
		versions = append(versions, seq)
	}
	if len(versions) == 0 {
		return 0, nil
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i] < versions[j] })
	return versions[len(versions)-1], nil
}
