package db

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"faster/internal/models"
)

// HasProfile reports whether onboarding has already been completed. The
// frontend calls this once on launch to decide whether to route to
// onboarding or straight into the app.
func (d *DB) HasProfile() (bool, error) {
	var count int
	err := d.Conn.QueryRow(`SELECT COUNT(*) FROM profile WHERE deleted_at IS NULL`).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetProfile returns the local user's profile, or nil if onboarding hasn't
// happened yet. This app supports a single active profile today; the
// profile_id FK on other tables leaves room for multi-profile later without
// a schema change.
func (d *DB) GetProfile() (*models.Profile, error) {
	row := d.Conn.QueryRow(`
		SELECT id, username, date_of_birth, weight_unit, height_unit, height_value, created_at, updated_at
		FROM profile WHERE deleted_at IS NULL
		ORDER BY created_at ASC LIMIT 1`)

	var p models.Profile
	err := row.Scan(&p.ID, &p.Username, &p.DateOfBirth, &p.WeightUnit, &p.HeightUnit, &p.HeightValue, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// CreateProfile runs onboarding as a single transaction: the profile row,
// and—if a starting weight was provided—its first weight_logs row. Doing
// both in one transaction avoids a partial-failure window where a profile
// exists but its first weight reading was silently dropped.
func (d *DB) CreateProfile(input models.CreateProfileInput) (models.Profile, error) {
	tx, err := d.Conn.Begin()
	if err != nil {
		return models.Profile{}, err
	}
	defer tx.Rollback()

	id := uuid.NewString()
	now := nowISO()

	_, err = tx.Exec(`
		INSERT INTO profile (id, username, date_of_birth, weight_unit, height_unit, height_value, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, input.Username, input.DateOfBirth, input.WeightUnit, input.HeightUnit, input.HeightValue, now, now,
	)
	if err != nil {
		return models.Profile{}, err
	}

	if input.WeightKg != nil {
		_, err = tx.Exec(`
			INSERT INTO weight_logs (id, profile_id, weight_kg, logged_at, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)`,
			uuid.NewString(), id, *input.WeightKg, now, now, now,
		)
		if err != nil {
			return models.Profile{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return models.Profile{}, err
	}

	return models.Profile{
		ID:          id,
		Username:    input.Username,
		DateOfBirth: input.DateOfBirth,
		WeightUnit:  input.WeightUnit,
		HeightUnit:  input.HeightUnit,
		HeightValue: input.HeightValue,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// UpdateProfile edits identity/unit fields (called from the Profile screen).
// Weight is intentionally not editable here — new weight readings go
// through AddWeightLog so history is preserved instead of overwritten.
func (d *DB) UpdateProfile(input models.UpdateProfileInput) (models.Profile, error) {
	existing, err := d.GetProfile()
	if err != nil {
		return models.Profile{}, err
	}
	if existing == nil {
		return models.Profile{}, ErrNotFound
	}

	now := nowISO()
	_, err = d.Conn.Exec(`
		UPDATE profile SET username=?, date_of_birth=?, weight_unit=?, height_unit=?, height_value=?, updated_at=?
		WHERE id=?`,
		input.Username, input.DateOfBirth, input.WeightUnit, input.HeightUnit, input.HeightValue, now, existing.ID,
	)
	if err != nil {
		return models.Profile{}, err
	}

	existing.Username = input.Username
	existing.DateOfBirth = input.DateOfBirth
	existing.WeightUnit = input.WeightUnit
	existing.HeightUnit = input.HeightUnit
	existing.HeightValue = input.HeightValue
	existing.UpdatedAt = now
	return *existing, nil
}
