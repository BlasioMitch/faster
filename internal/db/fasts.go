package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"faster/internal/models"
)

// ErrInvalidStartTime is returned by StartFast when a caller-supplied
// backdated start time is malformed, in the future, or further in the past
// than MaxBackdateHours allows.
var ErrInvalidStartTime = errors.New("invalid start time")

// MaxBackdateHours bounds how far in the past a fast's start time may be
// backdated to (see StartFast) — generous enough to cover "I forgot to open
// the app" for a day or more, while still rejecting obviously-wrong input.
const MaxBackdateHours = 7 * 24 // 7 days

// futureClockSkewTolerance absorbs small clock differences between the
// frontend's "now" (used to default/validate the start-time picker) and the
// backend's own clock, so a start time of "right now" is never spuriously
// rejected as "in the future".
const futureClockSkewTolerance = 2 * time.Minute

// StartFast begins a new fast. startTime is an RFC3339 timestamp for when
// the fast actually began; pass "" to mean "now" (the normal case). A
// non-empty value lets the user back-track a fast they forgot to start in
// the app at the time — see MaxBackdateHours for how far back that's
// allowed. The partial unique index idx_fasts_one_active enforces at the DB
// layer that a profile can have at most one fast with status='active' at a
// time; a violation is surfaced as ErrActiveFastExists rather than a raw
// SQL error.
func (d *DB) StartFast(profileID string, plannedDurationHours float64, startTime string) (models.Fast, error) {
	now := time.Now().UTC()

	start := now
	if startTime != "" {
		parsed, err := time.Parse(time.RFC3339, startTime)
		if err != nil {
			return models.Fast{}, fmt.Errorf("%w: %v", ErrInvalidStartTime, err)
		}
		parsed = parsed.UTC()
		if parsed.After(now.Add(futureClockSkewTolerance)) {
			return models.Fast{}, fmt.Errorf("%w: can't be in the future", ErrInvalidStartTime)
		}
		if parsed.Before(now.Add(-MaxBackdateHours * time.Hour)) {
			return models.Fast{}, fmt.Errorf("%w: can't be more than %d days in the past", ErrInvalidStartTime, MaxBackdateHours/24)
		}
		start = parsed
	}

	id := uuid.NewString()
	startStr := start.Format(time.RFC3339Nano)
	createdAt := nowISO() // when this record was created, distinct from the (possibly backdated) start_time

	_, err := d.Conn.Exec(`
		INSERT INTO fasts (id, profile_id, start_time, planned_duration_hours, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'active', ?, ?)`,
		id, profileID, startStr, plannedDurationHours, createdAt, createdAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return models.Fast{}, ErrActiveFastExists
		}
		return models.Fast{}, err
	}

	return models.Fast{
		ID:                   id,
		ProfileID:            profileID,
		StartTime:            startStr,
		PlannedDurationHours: plannedDurationHours,
		Status:               "active",
		CreatedAt:            createdAt,
		UpdatedAt:            createdAt,
	}, nil
}

// GetActiveFast returns the profile's in-progress fast, or nil if none.
// This is the only call the frontend needs on launch to rehydrate a
// countdown that was running when the app was last closed — no background
// ticking is stored or needed server-side.
func (d *DB) GetActiveFast(profileID string) (*models.Fast, error) {
	row := d.Conn.QueryRow(fastSelectColumns+`
		FROM fasts WHERE profile_id = ? AND status = 'active' AND deleted_at IS NULL LIMIT 1`, profileID)

	f, err := scanFastRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return f, nil
}

// CompleteFast ends a fast that reached (or passed) its planned duration.
func (d *DB) CompleteFast(id string, mood *int, note *string) (models.Fast, error) {
	return d.endFast(id, "completed", mood, note)
}

// AbandonFast ends a fast the user stopped before reaching its planned duration.
func (d *DB) AbandonFast(id string, mood *int, note *string) (models.Fast, error) {
	return d.endFast(id, "abandoned", mood, note)
}

func (d *DB) endFast(id string, status string, mood *int, note *string) (models.Fast, error) {
	tx, err := d.Conn.Begin()
	if err != nil {
		return models.Fast{}, err
	}
	defer tx.Rollback()

	var startTimeStr string
	err = tx.QueryRow(`SELECT start_time FROM fasts WHERE id = ? AND deleted_at IS NULL`, id).Scan(&startTimeStr)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Fast{}, ErrNotFound
	}
	if err != nil {
		return models.Fast{}, err
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return models.Fast{}, err
	}

	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)
	durationSeconds := int64(now.Sub(startTime).Seconds())
	if durationSeconds < 0 {
		durationSeconds = 0
	}

	_, err = tx.Exec(`
		UPDATE fasts SET end_time=?, actual_duration_seconds=?, status=?, mood=?, note=?, updated_at=?
		WHERE id=?`,
		nowStr, durationSeconds, status, mood, note, nowStr, id,
	)
	if err != nil {
		return models.Fast{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.Fast{}, err
	}

	row := d.Conn.QueryRow(fastSelectColumns+`FROM fasts WHERE id = ?`, id)
	f, err := scanFastRow(row)
	if err != nil {
		return models.Fast{}, err
	}
	return *f, nil
}

// ListFasts returns the profile's fast history, newest first. limit<=0
// defaults to 50.
func (d *DB) ListFasts(profileID string, limit, offset int) ([]models.Fast, error) {
	if limit <= 0 {
		limit = 50
	}

	// rowid (SQLite's implicit insertion-order column) is a tiebreaker for
	// start_time, so history stays deterministically ordered even if two
	// fasts were ever started within the same nanosecond-timestamp tick.
	rows, err := d.Conn.Query(fastSelectColumns+`
		FROM fasts WHERE profile_id = ? AND deleted_at IS NULL
		ORDER BY start_time DESC, rowid DESC LIMIT ? OFFSET ?`, profileID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fasts := []models.Fast{}
	for rows.Next() {
		f, err := scanFastRows(rows)
		if err != nil {
			return nil, err
		}
		fasts = append(fasts, *f)
	}
	return fasts, rows.Err()
}

const fastSelectColumns = `SELECT id, profile_id, start_time, planned_duration_hours, end_time, actual_duration_seconds, status, mood, note, created_at, updated_at `

func scanFastRow(row *sql.Row) (*models.Fast, error) {
	var f models.Fast
	err := row.Scan(&f.ID, &f.ProfileID, &f.StartTime, &f.PlannedDurationHours, &f.EndTime, &f.ActualDurationSeconds, &f.Status, &f.Mood, &f.Note, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func scanFastRows(rows *sql.Rows) (*models.Fast, error) {
	var f models.Fast
	err := rows.Scan(&f.ID, &f.ProfileID, &f.StartTime, &f.PlannedDurationHours, &f.EndTime, &f.ActualDurationSeconds, &f.Status, &f.Mood, &f.Note, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}
