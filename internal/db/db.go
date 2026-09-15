// Package db is the SQLite data-access layer. It uses modernc.org/sqlite
// (a pure-Go driver, no cgo) so the app's only cgo dependency remains the
// OS webview binding Wails itself needs — see the project plan for why.
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// ErrNotFound is returned when a lookup by id finds no row.
var ErrNotFound = errors.New("not found")

// ErrActiveFastExists is returned by StartFast when the profile already has
// an in-progress fast (enforced at the DB level by a partial unique index).
var ErrActiveFastExists = errors.New("a fast is already active")

// DB wraps the SQL connection. A single connection is intentional: this is
// a single-user desktop app with no concurrent-writer need, so capping at
// one connection sidesteps SQLITE_BUSY handling entirely.
type DB struct {
	Conn *sql.DB
}

// DefaultPath returns the production DB file location under the OS's
// per-user config directory (e.g. ~/.config/faster/faster.db on Linux),
// never cwd-relative.
func DefaultPath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	dir := filepath.Join(cfgDir, "faster")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create app config dir: %w", err)
	}
	return filepath.Join(dir, "faster.db"), nil
}

// Open opens (creating if needed) the SQLite database at path and runs any
// pending migrations. path may be a file path or ":memory:" for tests.
func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	conn.SetMaxOpenConns(1)

	pragmas := []string{
		`PRAGMA journal_mode = WAL;`,
		`PRAGMA synchronous = NORMAL;`,
		`PRAGMA foreign_keys = ON;`,
	}
	for _, p := range pragmas {
		if _, err := conn.Exec(p); err != nil {
			conn.Close()
			return nil, fmt.Errorf("set pragma %q: %w", p, err)
		}
	}

	d := &DB{Conn: conn}
	if err := d.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return d, nil
}

// Close closes the underlying connection.
func (d *DB) Close() error {
	return d.Conn.Close()
}

// nowISO returns the current instant as RFC3339 UTC with nanosecond
// precision — the timestamp format used for every
// created_at/updated_at/start_time/end_time column. Nanosecond precision
// (not plain RFC3339's whole-second precision) matters here: a fast can be
// started and completed within the same second (tests, or a user
// immediately re-starting after abandoning), and whole-second timestamps
// would make two such events indistinguishable, corrupting duration math
// and history ordering. time.Parse(time.RFC3339, ...) still reads these
// back fine — fractional seconds are always optional on parse regardless
// of the layout used to format them.
func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}
