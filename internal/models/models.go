// Package models holds the plain data structs shared between the Go backend
// and the frontend (via Wails JSON bindings). Every timestamp is an
// RFC3339 UTC string — kept as strings (not time.Time) across the Go<->JS
// boundary so the JSON contract is unambiguous on both sides.
package models

// Profile is the single local user's identity + body-metric preferences.
type Profile struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	DateOfBirth string   `json:"dateOfBirth"`           // 'YYYY-MM-DD'
	WeightUnit  string   `json:"weightUnit"`            // "kg" | "lb"
	HeightUnit  string   `json:"heightUnit"`            // "cm" | "ft_in"
	HeightValue *float64 `json:"heightValue,omitempty"` // canonical centimeters, optional
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

// CreateProfileInput is the onboarding form payload. WeightKg/HeightValue
// are optional per the product requirement; when WeightKg is present it
// becomes the first row in weight_logs, not a column on profile.
type CreateProfileInput struct {
	Username    string   `json:"username"`
	DateOfBirth string   `json:"dateOfBirth"`
	WeightUnit  string   `json:"weightUnit"`
	HeightUnit  string   `json:"heightUnit"`
	WeightKg    *float64 `json:"weightKg,omitempty"`    // canonical kg, optional
	HeightValue *float64 `json:"heightValue,omitempty"` // canonical cm, optional
}

// UpdateProfileInput edits identity/unit-preference fields from the Profile screen.
type UpdateProfileInput struct {
	Username    string   `json:"username"`
	DateOfBirth string   `json:"dateOfBirth"`
	WeightUnit  string   `json:"weightUnit"`
	HeightUnit  string   `json:"heightUnit"`
	HeightValue *float64 `json:"heightValue,omitempty"`
}

// WeightLog is one point-in-time weight measurement, always stored in
// canonical kilograms regardless of the profile's display unit preference.
type WeightLog struct {
	ID        string  `json:"id"`
	ProfileID string  `json:"profileId"`
	WeightKg  float64 `json:"weightKg"`
	LoggedAt  string  `json:"loggedAt"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

// Fast is one fasting session: active (in progress), completed (reached its
// goal), or abandoned (ended early). end_time/actual_duration_seconds stay
// nil until the fast reaches a terminal state.
type Fast struct {
	ID                    string  `json:"id"`
	ProfileID             string  `json:"profileId"`
	StartTime             string  `json:"startTime"`
	PlannedDurationHours  float64 `json:"plannedDurationHours"`
	EndTime               *string `json:"endTime,omitempty"`
	ActualDurationSeconds *int64  `json:"actualDurationSeconds,omitempty"`
	Status                string  `json:"status"`         // "active" | "completed" | "abandoned"
	Mood                  *int    `json:"mood,omitempty"` // 1..5, nullable
	Note                  *string `json:"note,omitempty"`
	CreatedAt             string  `json:"createdAt"`
	UpdatedAt             string  `json:"updatedAt"`
}

// Stats backs the Reports screen's summary cards.
type Stats struct {
	TotalFasts            int     `json:"totalFasts"`
	CompletionRatePercent float64 `json:"completionRatePercent"`
	LongestFastHours      float64 `json:"longestFastHours"`
	AverageDurationHours  float64 `json:"averageDurationHours"`
	CurrentStreakDays     int     `json:"currentStreakDays"`
}
