package db

import (
	"github.com/google/uuid"

	"faster/internal/models"
)

// AddWeightLog records a new weight reading. loggedAt is an RFC3339
// timestamp for the date/time the measurement corresponds to; if empty,
// "now" is used.
func (d *DB) AddWeightLog(profileID string, weightKg float64, loggedAt string) (models.WeightLog, error) {
	now := nowISO()
	if loggedAt == "" {
		loggedAt = now
	}
	id := uuid.NewString()

	_, err := d.Conn.Exec(`
		INSERT INTO weight_logs (id, profile_id, weight_kg, logged_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		id, profileID, weightKg, loggedAt, now, now,
	)
	if err != nil {
		return models.WeightLog{}, err
	}

	return models.WeightLog{
		ID:        id,
		ProfileID: profileID,
		WeightKg:  weightKg,
		LoggedAt:  loggedAt,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// ListWeightLogs returns all weight readings for the profile, oldest first,
// which is the order the Reports trend chart wants.
func (d *DB) ListWeightLogs(profileID string) ([]models.WeightLog, error) {
	rows, err := d.Conn.Query(`
		SELECT id, profile_id, weight_kg, logged_at, created_at, updated_at
		FROM weight_logs
		WHERE profile_id = ? AND deleted_at IS NULL
		ORDER BY logged_at ASC`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []models.WeightLog{}
	for rows.Next() {
		var w models.WeightLog
		if err := rows.Scan(&w.ID, &w.ProfileID, &w.WeightKg, &w.LoggedAt, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, w)
	}
	return logs, rows.Err()
}
