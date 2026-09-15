package db

import (
	"errors"
	"testing"
	"time"

	"faster/internal/models"
)

func newTestDB(t *testing.T) *DB {
	t.Helper()
	d, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open(:memory:): %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func mustCreateProfile(t *testing.T, d *DB) models.Profile {
	t.Helper()
	p, err := d.CreateProfile(models.CreateProfileInput{
		Username:    "alex",
		DateOfBirth: "1990-01-01",
		WeightUnit:  "kg",
		HeightUnit:  "cm",
	})
	if err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}
	return p
}

func TestCreateAndGetProfile(t *testing.T) {
	d := newTestDB(t)

	has, err := d.HasProfile()
	if err != nil {
		t.Fatalf("HasProfile: %v", err)
	}
	if has {
		t.Fatalf("HasProfile() = true before any profile created")
	}

	weight := 72.5
	height := 178.0
	p, err := d.CreateProfile(models.CreateProfileInput{
		Username:    "alex",
		DateOfBirth: "1990-01-01",
		WeightUnit:  "kg",
		HeightUnit:  "cm",
		WeightKg:    &weight,
		HeightValue: &height,
	})
	if err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}
	if p.ID == "" {
		t.Fatalf("CreateProfile returned empty ID")
	}

	has, err = d.HasProfile()
	if err != nil {
		t.Fatalf("HasProfile: %v", err)
	}
	if !has {
		t.Fatalf("HasProfile() = false after creating a profile")
	}

	got, err := d.GetProfile()
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if got == nil {
		t.Fatalf("GetProfile() = nil after creating a profile")
	}
	if got.Username != "alex" || got.HeightValue == nil || *got.HeightValue != 178.0 {
		t.Fatalf("GetProfile() = %+v, want username=alex height=178", got)
	}

	logs, err := d.ListWeightLogs(p.ID)
	if err != nil {
		t.Fatalf("ListWeightLogs: %v", err)
	}
	if len(logs) != 1 || logs[0].WeightKg != 72.5 {
		t.Fatalf("ListWeightLogs() = %+v, want one entry of 72.5kg (onboarding weight becomes first log)", logs)
	}
}

func TestCreateProfileWithoutOptionalFields(t *testing.T) {
	d := newTestDB(t)
	p := mustCreateProfile(t, d)

	if p.HeightValue != nil {
		t.Fatalf("HeightValue = %v, want nil when not provided", p.HeightValue)
	}
	logs, err := d.ListWeightLogs(p.ID)
	if err != nil {
		t.Fatalf("ListWeightLogs: %v", err)
	}
	if len(logs) != 0 {
		t.Fatalf("ListWeightLogs() = %+v, want none when no weight was given at onboarding", logs)
	}
}

func TestUpdateProfile(t *testing.T) {
	d := newTestDB(t)
	p := mustCreateProfile(t, d)

	newHeight := 180.0
	updated, err := d.UpdateProfile(models.UpdateProfileInput{
		Username:    "alexandra",
		DateOfBirth: p.DateOfBirth,
		WeightUnit:  "lb",
		HeightUnit:  "cm",
		HeightValue: &newHeight,
	})
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	if updated.Username != "alexandra" || updated.WeightUnit != "lb" || updated.HeightValue == nil || *updated.HeightValue != 180.0 {
		t.Fatalf("UpdateProfile() = %+v, unexpected result", updated)
	}
}

func TestStartFastEnforcesOneActive(t *testing.T) {
	d := newTestDB(t)
	p := mustCreateProfile(t, d)

	f, err := d.StartFast(p.ID, 24, "")
	if err != nil {
		t.Fatalf("StartFast: %v", err)
	}
	if f.Status != "active" {
		t.Fatalf("StartFast() status = %q, want active", f.Status)
	}

	_, err = d.StartFast(p.ID, 36, "")
	if err != ErrActiveFastExists {
		t.Fatalf("second StartFast() err = %v, want ErrActiveFastExists", err)
	}

	active, err := d.GetActiveFast(p.ID)
	if err != nil {
		t.Fatalf("GetActiveFast: %v", err)
	}
	if active == nil || active.ID != f.ID {
		t.Fatalf("GetActiveFast() = %+v, want the fast we started", active)
	}
}

func TestStartFastWithBackdatedStartTime(t *testing.T) {
	d := newTestDB(t)
	p := mustCreateProfile(t, d)

	startedAt := time.Now().UTC().Add(-3 * time.Hour)
	f, err := d.StartFast(p.ID, 24, startedAt.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("StartFast with backdated start: %v", err)
	}

	got, err := time.Parse(time.RFC3339Nano, f.StartTime)
	if err != nil {
		t.Fatalf("parse returned StartTime: %v", err)
	}
	if diff := got.Sub(startedAt); diff < -time.Second || diff > time.Second {
		t.Fatalf("StartTime = %v, want ~%v (the backdated time we supplied)", got, startedAt)
	}
}

func TestStartFastRejectsFutureStartTime(t *testing.T) {
	d := newTestDB(t)
	p := mustCreateProfile(t, d)

	future := time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339)
	_, err := d.StartFast(p.ID, 24, future)
	if !errors.Is(err, ErrInvalidStartTime) {
		t.Fatalf("StartFast with future start time: err = %v, want ErrInvalidStartTime", err)
	}
}

func TestStartFastRejectsExcessiveBackdate(t *testing.T) {
	d := newTestDB(t)
	p := mustCreateProfile(t, d)

	tooFarBack := time.Now().UTC().Add(-(MaxBackdateHours + 1) * time.Hour).Format(time.RFC3339)
	_, err := d.StartFast(p.ID, 24, tooFarBack)
	if !errors.Is(err, ErrInvalidStartTime) {
		t.Fatalf("StartFast beyond MaxBackdateHours: err = %v, want ErrInvalidStartTime", err)
	}
}

func TestStartFastRejectsMalformedStartTime(t *testing.T) {
	d := newTestDB(t)
	p := mustCreateProfile(t, d)

	_, err := d.StartFast(p.ID, 24, "not-a-timestamp")
	if !errors.Is(err, ErrInvalidStartTime) {
		t.Fatalf("StartFast with malformed start time: err = %v, want ErrInvalidStartTime", err)
	}
}

func TestCompleteFastRecordsMoodAndDuration(t *testing.T) {
	d := newTestDB(t)
	p := mustCreateProfile(t, d)

	f, err := d.StartFast(p.ID, 24, "")
	if err != nil {
		t.Fatalf("StartFast: %v", err)
	}

	mood := 5
	note := "felt great"
	done, err := d.CompleteFast(f.ID, &mood, &note)
	if err != nil {
		t.Fatalf("CompleteFast: %v", err)
	}
	if done.Status != "completed" {
		t.Fatalf("CompleteFast() status = %q, want completed", done.Status)
	}
	if done.Mood == nil || *done.Mood != 5 {
		t.Fatalf("CompleteFast() mood = %v, want 5", done.Mood)
	}
	if done.Note == nil || *done.Note != "felt great" {
		t.Fatalf("CompleteFast() note = %v, want %q", done.Note, "felt great")
	}
	if done.ActualDurationSeconds == nil {
		t.Fatalf("CompleteFast() ActualDurationSeconds is nil, want set")
	}
	if done.EndTime == nil {
		t.Fatalf("CompleteFast() EndTime is nil, want set")
	}

	// Starting a new fast should now succeed since none is active.
	if _, err := d.StartFast(p.ID, 36, ""); err != nil {
		t.Fatalf("StartFast after completion: %v", err)
	}
}

func TestAbandonFast(t *testing.T) {
	d := newTestDB(t)
	p := mustCreateProfile(t, d)

	f, err := d.StartFast(p.ID, 48, "")
	if err != nil {
		t.Fatalf("StartFast: %v", err)
	}

	mood := 2
	done, err := d.AbandonFast(f.ID, &mood, nil)
	if err != nil {
		t.Fatalf("AbandonFast: %v", err)
	}
	if done.Status != "abandoned" {
		t.Fatalf("AbandonFast() status = %q, want abandoned", done.Status)
	}
}

func TestListFastsOrderingAndPagination(t *testing.T) {
	d := newTestDB(t)
	p := mustCreateProfile(t, d)

	var ids []string
	for i := 0; i < 3; i++ {
		f, err := d.StartFast(p.ID, 24, "")
		if err != nil {
			t.Fatalf("StartFast #%d: %v", i, err)
		}
		if _, err := d.CompleteFast(f.ID, nil, nil); err != nil {
			t.Fatalf("CompleteFast #%d: %v", i, err)
		}
		ids = append(ids, f.ID)
	}

	all, err := d.ListFasts(p.ID, 0, 0)
	if err != nil {
		t.Fatalf("ListFasts: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("ListFasts() returned %d fasts, want 3", len(all))
	}
	// Newest first: the last-started fast (ids[2]) should come first.
	if all[0].ID != ids[2] {
		t.Fatalf("ListFasts()[0].ID = %s, want most recently started fast %s", all[0].ID, ids[2])
	}

	page, err := d.ListFasts(p.ID, 1, 1)
	if err != nil {
		t.Fatalf("ListFasts with pagination: %v", err)
	}
	if len(page) != 1 || page[0].ID != all[1].ID {
		t.Fatalf("ListFasts(limit=1, offset=1) = %+v, want just %+v", page, all[1])
	}
}

// backdateFastStart moves a fast's start_time into the past, so tests can
// produce a deterministic non-trivial actual_duration_seconds on
// completion without sleeping in real wall-clock time.
func backdateFastStart(t *testing.T, d *DB, fastID string, hoursAgo float64) {
	t.Helper()
	past := time.Now().UTC().Add(-time.Duration(hoursAgo * float64(time.Hour))).Format(time.RFC3339Nano)
	if _, err := d.Conn.Exec(`UPDATE fasts SET start_time = ? WHERE id = ?`, past, fastID); err != nil {
		t.Fatalf("backdateFastStart: %v", err)
	}
}

func TestGetStats(t *testing.T) {
	d := newTestDB(t)
	p := mustCreateProfile(t, d)

	f1, _ := d.StartFast(p.ID, 24, "")
	backdateFastStart(t, d, f1.ID, 24) // simulate a fast that actually ran ~24h
	if _, err := d.CompleteFast(f1.ID, nil, nil); err != nil {
		t.Fatalf("CompleteFast f1: %v", err)
	}

	f2, _ := d.StartFast(p.ID, 24, "")
	if _, err := d.AbandonFast(f2.ID, nil, nil); err != nil {
		t.Fatalf("AbandonFast f2: %v", err)
	}

	stats, err := d.GetStats(p.ID)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.TotalFasts != 2 {
		t.Fatalf("TotalFasts = %d, want 2", stats.TotalFasts)
	}
	if stats.CompletionRatePercent != 50 {
		t.Fatalf("CompletionRatePercent = %v, want 50", stats.CompletionRatePercent)
	}
	// Should equal the one completed fast's ~24h duration, not be dragged
	// toward 0 by the abandoned fast (which had no real elapsed time).
	if stats.AverageDurationHours < 23.9 || stats.AverageDurationHours > 24.1 {
		t.Fatalf("AverageDurationHours = %v, want ~24 (only the completed fast should count)", stats.AverageDurationHours)
	}
}

func TestGetStatsWithNoFasts(t *testing.T) {
	d := newTestDB(t)
	p := mustCreateProfile(t, d)

	stats, err := d.GetStats(p.ID)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.TotalFasts != 0 || stats.CompletionRatePercent != 0 || stats.CurrentStreakDays != 0 {
		t.Fatalf("GetStats() with no fasts = %+v, want all zero", stats)
	}
}

func TestAddAndListWeightLogs(t *testing.T) {
	d := newTestDB(t)
	p := mustCreateProfile(t, d)

	if _, err := d.AddWeightLog(p.ID, 70.0, "2024-01-01T00:00:00Z"); err != nil {
		t.Fatalf("AddWeightLog #1: %v", err)
	}
	if _, err := d.AddWeightLog(p.ID, 69.0, "2024-01-15T00:00:00Z"); err != nil {
		t.Fatalf("AddWeightLog #2: %v", err)
	}

	logs, err := d.ListWeightLogs(p.ID)
	if err != nil {
		t.Fatalf("ListWeightLogs: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("ListWeightLogs() len = %d, want 2", len(logs))
	}
	if logs[0].LoggedAt > logs[1].LoggedAt {
		t.Fatalf("ListWeightLogs() not ordered oldest-first: %+v", logs)
	}
}

// mustCompletedFastOnDaysAgo creates a completed fast whose start_time falls
// on local-calendar-day (today - daysAgo), regardless of what time of day
// the test itself runs at (anchored to local noon to avoid midnight edge
// cases), for exercising the streak calculation deterministically.
func mustCompletedFastOnDaysAgo(t *testing.T, d *DB, profileID string, daysAgo int) {
	t.Helper()
	f, err := d.StartFast(profileID, 24, "")
	if err != nil {
		t.Fatalf("StartFast: %v", err)
	}
	day := time.Now().Local().AddDate(0, 0, -daysAgo)
	noon := time.Date(day.Year(), day.Month(), day.Day(), 12, 0, 0, 0, day.Location())
	if _, err := d.Conn.Exec(`UPDATE fasts SET start_time = ? WHERE id = ?`, noon.UTC().Format(time.RFC3339Nano), f.ID); err != nil {
		t.Fatalf("backdate: %v", err)
	}
	if _, err := d.CompleteFast(f.ID, nil, nil); err != nil {
		t.Fatalf("CompleteFast: %v", err)
	}
}

func TestCurrentStreakConsecutiveDays(t *testing.T) {
	d := newTestDB(t)
	p := mustCreateProfile(t, d)

	mustCompletedFastOnDaysAgo(t, d, p.ID, 0)
	mustCompletedFastOnDaysAgo(t, d, p.ID, 1)
	mustCompletedFastOnDaysAgo(t, d, p.ID, 2)

	stats, err := d.GetStats(p.ID)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.CurrentStreakDays != 3 {
		t.Fatalf("CurrentStreakDays = %d, want 3", stats.CurrentStreakDays)
	}
}

func TestCurrentStreakGapBreaksIt(t *testing.T) {
	d := newTestDB(t)
	p := mustCreateProfile(t, d)

	mustCompletedFastOnDaysAgo(t, d, p.ID, 0) // today
	// yesterday (daysAgo=1) intentionally skipped
	mustCompletedFastOnDaysAgo(t, d, p.ID, 2)

	stats, err := d.GetStats(p.ID)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.CurrentStreakDays != 1 {
		t.Fatalf("CurrentStreakDays = %d, want 1 (gap at yesterday should stop the streak)", stats.CurrentStreakDays)
	}
}

func TestCurrentStreakNoFastTodayContinuesFromYesterday(t *testing.T) {
	d := newTestDB(t)
	p := mustCreateProfile(t, d)

	// Nothing completed today (e.g. an active fast in progress), but
	// yesterday and the day before both have completed fasts.
	mustCompletedFastOnDaysAgo(t, d, p.ID, 1)
	mustCompletedFastOnDaysAgo(t, d, p.ID, 2)

	stats, err := d.GetStats(p.ID)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.CurrentStreakDays != 2 {
		t.Fatalf("CurrentStreakDays = %d, want 2 (today having no completed fast yet shouldn't break yesterday's streak)", stats.CurrentStreakDays)
	}
}

func TestCurrentStreakAbandonedDoesNotCount(t *testing.T) {
	d := newTestDB(t)
	p := mustCreateProfile(t, d)

	f, err := d.StartFast(p.ID, 24, "")
	if err != nil {
		t.Fatalf("StartFast: %v", err)
	}
	if _, err := d.AbandonFast(f.ID, nil, nil); err != nil {
		t.Fatalf("AbandonFast: %v", err)
	}

	stats, err := d.GetStats(p.ID)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.CurrentStreakDays != 0 {
		t.Fatalf("CurrentStreakDays = %d, want 0 (an abandoned fast must not count toward a streak)", stats.CurrentStreakDays)
	}
}
