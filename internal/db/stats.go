package db

import (
	"database/sql"
	"time"

	"faster/internal/models"
)

// GetStats computes the Reports screen's summary cards from scratch on
// every call. At the data volumes this app will ever see (a handful of
// fasts a week), that's simpler and cheap enough to not need caching.
func (d *DB) GetStats(profileID string) (models.Stats, error) {
	var stats models.Stats

	var completed, abandoned int
	if err := d.Conn.QueryRow(
		`SELECT COUNT(*) FROM fasts WHERE profile_id=? AND status='completed' AND deleted_at IS NULL`,
		profileID,
	).Scan(&completed); err != nil {
		return stats, err
	}
	if err := d.Conn.QueryRow(
		`SELECT COUNT(*) FROM fasts WHERE profile_id=? AND status='abandoned' AND deleted_at IS NULL`,
		profileID,
	).Scan(&abandoned); err != nil {
		return stats, err
	}

	stats.TotalFasts = completed + abandoned
	if stats.TotalFasts > 0 {
		stats.CompletionRatePercent = float64(completed) / float64(stats.TotalFasts) * 100
	}

	var longestSeconds, avgSeconds sql.NullFloat64
	if err := d.Conn.QueryRow(
		`SELECT MAX(actual_duration_seconds) FROM fasts WHERE profile_id=? AND status='completed' AND deleted_at IS NULL`,
		profileID,
	).Scan(&longestSeconds); err != nil {
		return stats, err
	}
	if longestSeconds.Valid {
		stats.LongestFastHours = longestSeconds.Float64 / 3600
	}

	// Average only counts completed fasts, not abandoned ones — this keeps
	// "average duration" meaning "typical successful fast length" rather
	// than being dragged down by early-ended attempts (which already show
	// up in CompletionRatePercent).
	if err := d.Conn.QueryRow(
		`SELECT AVG(actual_duration_seconds) FROM fasts WHERE profile_id=? AND status='completed' AND deleted_at IS NULL`,
		profileID,
	).Scan(&avgSeconds); err != nil {
		return stats, err
	}
	if avgSeconds.Valid {
		stats.AverageDurationHours = avgSeconds.Float64 / 3600
	}

	streak, err := d.currentStreak(profileID)
	if err != nil {
		return stats, err
	}
	stats.CurrentStreakDays = streak

	return stats, nil
}

// currentStreak counts consecutive local-calendar days with at least one
// completed fast, walking backward from today. If today has no completed
// fast yet (e.g. a fast is still active), we start from yesterday instead —
// otherwise a same-day in-progress fast would look like it broke yesterday's
// streak. Abandoned fasts never count toward a streak.
func (d *DB) currentStreak(profileID string) (int, error) {
	rows, err := d.Conn.Query(
		`SELECT start_time FROM fasts WHERE profile_id=? AND status='completed' AND deleted_at IS NULL ORDER BY start_time DESC`,
		profileID,
	)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	dateSet := map[string]bool{}
	for rows.Next() {
		var startStr string
		if err := rows.Scan(&startStr); err != nil {
			return 0, err
		}
		t, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			continue
		}
		dateSet[t.Local().Format("2006-01-02")] = true
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if len(dateSet) == 0 {
		return 0, nil
	}

	cursor := time.Now().Local()
	if !dateSet[cursor.Format("2006-01-02")] {
		cursor = cursor.AddDate(0, 0, -1)
	}

	streak := 0
	for dateSet[cursor.Format("2006-01-02")] {
		streak++
		cursor = cursor.AddDate(0, 0, -1)
	}
	return streak, nil
}
