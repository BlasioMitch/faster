package db

import (
	"database/sql"
	"fmt"
)

// migrations runs in order, gated by PRAGMA user_version. Each entry is
// applied at most once, in its own transaction. Add new entries to the end
// of this slice for future schema changes — never edit an already-shipped
// entry, since that would desync existing users' databases.
var migrations = []func(*sql.Tx) error{
	migration001InitialSchema,
}

func (d *DB) migrate() error {
	var version int
	if err := d.Conn.QueryRow(`PRAGMA user_version;`).Scan(&version); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}

	for i := version; i < len(migrations); i++ {
		tx, err := d.Conn.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", i+1, err)
		}
		if err := migrations[i](tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %d: %w", i+1, err)
		}
		// PRAGMA user_version does not support bound parameters in SQLite;
		// i+1 is an internally-controlled int, not user input.
		if _, err := tx.Exec(fmt.Sprintf(`PRAGMA user_version = %d;`, i+1)); err != nil {
			tx.Rollback()
			return fmt.Errorf("bump schema version to %d: %w", i+1, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", i+1, err)
		}
	}
	return nil
}

// migration001InitialSchema creates the full initial schema: profile,
// weight_logs, fasts. Every table carries a UUID text primary key plus
// created_at/updated_at/deleted_at so a future sync engine can diff
// "updated_at > cursor" and treat deleted_at as a tombstone, without any
// schema change needed later.
func migration001InitialSchema(tx *sql.Tx) error {
	stmts := []string{
		`CREATE TABLE profile (
			id            TEXT PRIMARY KEY,
			username      TEXT NOT NULL,
			date_of_birth TEXT NOT NULL,
			weight_unit   TEXT NOT NULL DEFAULT 'kg',
			height_unit   TEXT NOT NULL DEFAULT 'cm',
			height_value  REAL,
			created_at    TEXT NOT NULL,
			updated_at    TEXT NOT NULL,
			deleted_at    TEXT
		);`,
		`CREATE TABLE weight_logs (
			id         TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL REFERENCES profile(id),
			weight_kg  REAL NOT NULL,
			logged_at  TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			deleted_at TEXT
		);`,
		`CREATE INDEX idx_weight_logs_profile_time ON weight_logs(profile_id, logged_at);`,
		`CREATE TABLE fasts (
			id                       TEXT PRIMARY KEY,
			profile_id               TEXT NOT NULL REFERENCES profile(id),
			start_time               TEXT NOT NULL,
			planned_duration_hours   REAL NOT NULL,
			end_time                 TEXT,
			actual_duration_seconds  INTEGER,
			status                   TEXT NOT NULL DEFAULT 'active',
			mood                     INTEGER,
			note                     TEXT,
			created_at               TEXT NOT NULL,
			updated_at               TEXT NOT NULL,
			deleted_at               TEXT
		);`,
		`CREATE INDEX idx_fasts_profile_start ON fasts(profile_id, start_time DESC);`,
		// Enforces "only one truly in-progress fast per profile" at the DB
		// layer, not just in application logic.
		`CREATE UNIQUE INDEX idx_fasts_one_active
			ON fasts(profile_id) WHERE status = 'active' AND deleted_at IS NULL;`,
	}
	for _, s := range stmts {
		if _, err := tx.Exec(s); err != nil {
			return fmt.Errorf("exec %q: %w", s, err)
		}
	}
	return nil
}
