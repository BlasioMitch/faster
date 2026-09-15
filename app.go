package main

import (
	"context"
	"errors"
	"fmt"

	"faster/internal/db"
	"faster/internal/models"
)

// App is the struct whose exported methods Wails binds to the frontend.
// It holds no state of its own beyond the DB handle and Wails context —
// every screen's data is read fresh from SQLite on each call.
type App struct {
	ctx context.Context
	db  *db.DB
}

// NewApp creates a new App backed by the given already-opened database.
// The DB connection lifecycle (open at process start, close at shutdown)
// is managed in main.go, not here.
func NewApp(database *db.DB) *App {
	return &App{db: database}
}

// startup is called when the Wails app starts; the context lets us call
// runtime methods later if ever needed (e.g. native dialogs).
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// currentProfileID is a small internal helper: nearly every fasting/report
// method needs "the" profile id, and this app supports exactly one local
// profile today (see the plan's multi-profile-readiness note on the schema).
func (a *App) currentProfileID() (string, error) {
	p, err := a.db.GetProfile()
	if err != nil {
		return "", err
	}
	if p == nil {
		return "", errors.New("no profile exists yet — call CreateProfile first")
	}
	return p.ID, nil
}

// --- Profile -----------------------------------------------------------

// HasProfile reports whether onboarding has been completed.
func (a *App) HasProfile() (bool, error) {
	return a.db.HasProfile()
}

// GetProfile returns the local profile, or nil before onboarding.
func (a *App) GetProfile() (*models.Profile, error) {
	return a.db.GetProfile()
}

// CreateProfile completes onboarding.
func (a *App) CreateProfile(input models.CreateProfileInput) (models.Profile, error) {
	return a.db.CreateProfile(input)
}

// UpdateProfile edits identity/unit-preference fields from the Profile screen.
func (a *App) UpdateProfile(input models.UpdateProfileInput) (models.Profile, error) {
	return a.db.UpdateProfile(input)
}

// AddWeightLog records a new weight reading for the current profile.
func (a *App) AddWeightLog(weightKg float64, loggedAt string) (models.WeightLog, error) {
	profileID, err := a.currentProfileID()
	if err != nil {
		return models.WeightLog{}, err
	}
	return a.db.AddWeightLog(profileID, weightKg, loggedAt)
}

// ListWeightLogs returns the current profile's weight history, oldest first.
func (a *App) ListWeightLogs() ([]models.WeightLog, error) {
	profileID, err := a.currentProfileID()
	if err != nil {
		return nil, err
	}
	return a.db.ListWeightLogs(profileID)
}

// --- Fasting -------------------------------------------------------------

// GetActiveFast returns the in-progress fast, or nil if none. The frontend
// calls this once on launch to rehydrate a countdown that was running when
// the app was last closed.
func (a *App) GetActiveFast() (*models.Fast, error) {
	profileID, err := a.currentProfileID()
	if err != nil {
		return nil, err
	}
	return a.db.GetActiveFast(profileID)
}

// StartFast begins a new fast for the given planned duration (hours; a
// preset like 24/36/48/72/120, or a custom value). startTime is an RFC3339
// timestamp for when the fast actually began — pass "" to mean "now", or a
// past timestamp to back-track a fast the user forgot to start in the app
// at the time (bounded by db.MaxBackdateHours). Fails with a descriptive
// error if a fast is already active, or if startTime is invalid.
func (a *App) StartFast(plannedDurationHours float64, startTime string) (models.Fast, error) {
	profileID, err := a.currentProfileID()
	if err != nil {
		return models.Fast{}, err
	}
	f, err := a.db.StartFast(profileID, plannedDurationHours, startTime)
	if errors.Is(err, db.ErrActiveFastExists) {
		return models.Fast{}, fmt.Errorf("a fast is already in progress")
	}
	if errors.Is(err, db.ErrInvalidStartTime) {
		return models.Fast{}, err
	}
	return f, err
}

// CompleteFast ends a fast that reached its goal. mood is 1..5 or 0 for
// "not set"; note may be empty.
func (a *App) CompleteFast(id string, mood int, note string) (models.Fast, error) {
	return a.db.CompleteFast(id, moodPtr(mood), notePtr(note))
}

// AbandonFast ends a fast the user stopped before reaching its goal.
func (a *App) AbandonFast(id string, mood int, note string) (models.Fast, error) {
	return a.db.AbandonFast(id, moodPtr(mood), notePtr(note))
}

// ListFasts returns fast history for the current profile, newest first.
func (a *App) ListFasts(limit, offset int) ([]models.Fast, error) {
	profileID, err := a.currentProfileID()
	if err != nil {
		return nil, err
	}
	return a.db.ListFasts(profileID, limit, offset)
}

// --- Reports ---------------------------------------------------------------

// GetStats returns the summary stats for the Reports screen.
func (a *App) GetStats() (models.Stats, error) {
	profileID, err := a.currentProfileID()
	if err != nil {
		return models.Stats{}, err
	}
	return a.db.GetStats(profileID)
}

// moodPtr/notePtr translate the JS-friendly "0/empty means unset" convention
// (Wails' JSON bridge makes plain optional int/string params awkward) into
// the nullable *int/*string the DB layer expects.
func moodPtr(mood int) *int {
	if mood <= 0 {
		return nil
	}
	return &mood
}

func notePtr(note string) *string {
	if note == "" {
		return nil
	}
	return &note
}
