package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// migrationLockName is the advisory lock name used to prevent concurrent migrations.
const migrationLockName = "lattice_migrations"

// MigrateUp applies all pending up-migrations from the given directory.
// The database DSN must include multiStatements=true for migrations containing
// multiple SQL statements.
//
// Note: MySQL implicitly commits DDL statements, so transactions cannot roll back
// DDL. Each migration file should contain a single DDL operation to avoid partial
// application. The version is recorded after the DDL executes successfully.
//
// Acquires a MySQL advisory lock on a dedicated connection to prevent concurrent
// migration execution.
func MigrateUp(ctx context.Context, db *sql.DB, dir string) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if err := acquireMigrationLock(ctx, conn); err != nil {
		return err
	}
	defer releaseMigrationLock(conn)

	if err := ensureMigrationsTable(ctx, conn); err != nil {
		return fmt.Errorf("ensure migrations table: %w", err)
	}

	applied, err := appliedVersions(ctx, conn)
	if err != nil {
		return fmt.Errorf("read applied versions: %w", err)
	}

	files, err := migrationFiles(dir, ".up.sql")
	if err != nil {
		return err
	}

	for _, f := range files {
		ver, err := extractVersion(f)
		if err != nil {
			return fmt.Errorf("parse migration version %s: %w", f, err)
		}
		if applied[ver] {
			continue
		}
		if err := applyMigration(ctx, conn, dir, f, ver); err != nil {
			return err
		}
	}
	return nil
}

func applyMigration(ctx context.Context, conn *sql.Conn, dir, filename string, ver int) error {
	content, err := os.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		return fmt.Errorf("read migration %s: %w", filename, err)
	}

	// DDL in MySQL causes implicit commit, so we execute the migration SQL
	// directly and record the version separately. If recording fails after
	// DDL succeeds, the migration must be resolved manually.
	if _, err := conn.ExecContext(ctx, string(content)); err != nil {
		return fmt.Errorf("apply migration %s: %w", filename, err)
	}
	if _, err := conn.ExecContext(ctx,
		"INSERT INTO schema_migrations (version) VALUES (?)", ver); err != nil {
		return fmt.Errorf("record migration %s: %w", filename, err)
	}
	return nil
}

// MigrateDown rolls back the last n applied migrations.
// The database DSN must include multiStatements=true for migrations containing
// multiple SQL statements.
//
// Note: Rollbacks are applied one at a time. If a rollback fails midway through
// a multi-step rollback, the database will be in a partially rolled-back state
// and requires manual intervention.
//
// Acquires a MySQL advisory lock on a dedicated connection to prevent concurrent
// migration execution.
func MigrateDown(ctx context.Context, db *sql.DB, dir string, n int) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if err := acquireMigrationLock(ctx, conn); err != nil {
		return err
	}
	defer releaseMigrationLock(conn)

	if err := ensureMigrationsTable(ctx, conn); err != nil {
		return fmt.Errorf("ensure migrations table: %w", err)
	}

	versions, err := appliedVersionsOrdered(ctx, conn)
	if err != nil {
		return err
	}

	if n > len(versions) {
		n = len(versions)
	}

	downFiles, err := migrationFiles(dir, ".down.sql")
	if err != nil {
		return err
	}

	// Build a lookup from version to down-file.
	downByVersion := make(map[int]string, len(downFiles))
	for _, f := range downFiles {
		v, err := extractVersion(f)
		if err != nil {
			return fmt.Errorf("parse migration version %s: %w", f, err)
		}
		downByVersion[v] = f
	}

	// Rollback in reverse order (most recent first).
	for i := len(versions) - 1; i >= len(versions)-n; i-- {
		ver := versions[i]
		downFile, ok := downByVersion[ver]
		if !ok {
			return fmt.Errorf("no down migration found for version %03d", ver)
		}
		if err := rollbackMigration(ctx, conn, dir, downFile, ver); err != nil {
			return err
		}
	}
	return nil
}

func rollbackMigration(ctx context.Context, conn *sql.Conn, dir, filename string, ver int) error {
	content, err := os.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		return fmt.Errorf("read migration %s: %w", filename, err)
	}

	if _, err := conn.ExecContext(ctx, string(content)); err != nil {
		return fmt.Errorf("rollback migration %s: %w", filename, err)
	}
	if _, err := conn.ExecContext(ctx,
		"DELETE FROM schema_migrations WHERE version = ?", ver); err != nil {
		return fmt.Errorf("remove migration record %d: %w", ver, err)
	}
	return nil
}

func acquireMigrationLock(ctx context.Context, conn *sql.Conn) error {
	var result sql.NullInt64
	err := conn.QueryRowContext(ctx,
		"SELECT GET_LOCK(?, 10)", migrationLockName).Scan(&result)
	if err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	if !result.Valid || result.Int64 != 1 {
		return fmt.Errorf("could not acquire migration lock (timeout)")
	}
	return nil
}

func releaseMigrationLock(conn *sql.Conn) {
	// Use a background context with a short timeout to ensure the lock is
	// released even if the original migration context was canceled.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, _ = conn.ExecContext(ctx, "SELECT RELEASE_LOCK(?)", migrationLockName)
}

func ensureMigrationsTable(ctx context.Context, conn *sql.Conn) error {
	_, err := conn.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INT NOT NULL PRIMARY KEY,
			applied_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
		)`)
	return err
}

func appliedVersions(ctx context.Context, conn *sql.Conn) (map[int]bool, error) {
	rows, err := conn.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	result := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		result[v] = true
	}
	return result, rows.Err()
}

func appliedVersionsOrdered(ctx context.Context, conn *sql.Conn) ([]int, error) {
	rows, err := conn.QueryContext(ctx, "SELECT version FROM schema_migrations ORDER BY version ASC")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []int
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

type versionedFile struct {
	name    string
	version int
}

func migrationFiles(dir, suffix string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}
	var vfiles []versionedFile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), suffix) {
			continue
		}
		ver, err := extractVersion(e.Name())
		if err != nil {
			return nil, fmt.Errorf("parse migration file %s: %w", e.Name(), err)
		}
		vfiles = append(vfiles, versionedFile{name: e.Name(), version: ver})
	}
	sort.Slice(vfiles, func(i, j int) bool {
		return vfiles[i].version < vfiles[j].version
	})
	files := make([]string, len(vfiles))
	for i, vf := range vfiles {
		files[i] = vf.name
	}
	return files, nil
}

func extractVersion(filename string) (int, error) {
	parts := strings.SplitN(filename, "_", 2)
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid migration filename: %s", filename)
	}
	v, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid migration version in %s: %w", filename, err)
	}
	return v, nil
}
