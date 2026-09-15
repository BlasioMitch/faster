# Faster — Offline Fasting Tracker (Desktop, Wails v2)

## Context

The user wants a desktop app to track intermittent/extended fasting habits, with mobile to follow later. The repo (`/home/zeus/Documents/Projects/faster`) is currently empty — this is a from-scratch build. Core intent: a private, offline-first, no-login habit tracker that lets someone pick a fasting duration, watch a live circular countdown (with light educational context on what's happening physiologically at each stage), and later review their history and stats. Cross-device sync was mentioned but is explicitly deferred — this phase only needs to leave the data model sync-friendly, not implement sync.

**Decisions already confirmed with the user** (do not re-litigate during implementation):
- **Wails v2** (stable), not the v3 alpha.
- **Vanilla TypeScript** frontend (no React/Vue/Svelte) — smallest memory/bundle footprint.
- **Sync**: not built in this phase. Schema only needs to be sync-ready (UUID PKs, timestamps, soft-delete).
- **Sentiment**: a 5-point mood scale (Struggled → Great) plus an optional free-text note, captured when a fast ends.

## Environment setup (blocking, must happen first)

This machine is missing WebKitGTK dev headers and the Wails CLI:
```bash
sudo apt update
sudo apt install -y libwebkit2gtk-4.1-dev build-essential pkg-config
go install github.com/wailsapp/wails/v2/cmd/wails@latest
export PATH=$PATH:$(go env GOPATH)/bin   # persist in ~/.bashrc
wails doctor                              # confirm all green before scaffolding
```
`wails init -t vanilla-ts` creates its own directory, so scaffold into a sibling folder and merge into the existing (git-initialized, empty) `faster/` repo rather than targeting it directly.

## Tech stack & key architecture decisions

- **SQLite driver: `modernc.org/sqlite`** (pure Go, no cgo) over `mattn/go-sqlite3`. Wails already needs OS-specific cgo for the webview; a pure-Go DB driver avoids a *second* cgo dependency, which matters most when this app later cross-builds for Windows/macOS. `db.SetMaxOpenConns(1)` (single-user app, sidesteps `SQLITE_BUSY` entirely) + WAL journal mode.
- **DB location**: `os.UserConfigDir()/faster/faster.db`, not cwd-relative.
- **No backend ticking goroutine for the timer.** `StartFast` returns `start_time` + `planned_duration_hours`; the frontend runs a 1s `setInterval` that always recomputes `remaining = targetEnd - Date.now()` (wall-clock, self-correcting, no drift). Go is only called at 4 points: app launch (`GetActiveFast`, to rehydrate an in-progress fast after restart), `StartFast`, `CompleteFast`, `AbandonFast`. Between those the Go process is fully idle — this is the main lever for the low-memory goal on the backend side.
- **No chart/animation libraries.** Circular timer is a hand-rolled SVG ring (`stroke-dasharray`/`stroke-dashoffset`), updated once per second; CSS `transition: 1s linear` lets the compositor interpolate smoothly without a `requestAnimationFrame` loop. Reports' weight/mood trend charts are a small hand-rolled SVG polyline component (~100 lines), not a charting library.

## Data model (SQLite)

Every table: UUID `TEXT` primary key, `created_at`/`updated_at` (ISO-8601 UTC), `deleted_at` (nullable soft-delete) — this is the full extent of "sync-ready" for this phase; no change-log/outbox table yet (cheap to add later, not needed now).

```sql
CREATE TABLE profile (
  id, username, date_of_birth,           -- age is derived at read time, never stored
  weight_unit ('kg'|'lb'), height_unit ('cm'|'ft_in'),
  height_value REAL,                     -- nullable; canonical cm regardless of display unit
  created_at, updated_at, deleted_at
);

CREATE TABLE weight_logs (               -- separate table, not a column on profile
  id, profile_id, weight_kg REAL,        -- canonical kg; onboarding's weight IS the first row here
  logged_at, created_at, updated_at, deleted_at
);

CREATE TABLE fasts (
  id, profile_id,
  start_time, planned_duration_hours REAL,   -- accepts presets or custom values
  end_time, actual_duration_seconds,         -- nullable until terminal
  status ('active'|'completed'|'abandoned'), -- 'active' needed for rehydration + the unique-index below
  mood INTEGER (1..5, nullable), note (nullable),
  created_at, updated_at, deleted_at
);
CREATE UNIQUE INDEX idx_fasts_one_active ON fasts(profile_id)
  WHERE status='active' AND deleted_at IS NULL;   -- enforces only one in-progress fast
```
Fasting-duration presets (24/36/48/72/120h + a custom 1–240h option) and the fasting-stage educational content are **static frontend data**, not DB tables — identical for every user, no query benefit from a table.

## Go ↔ JS binding surface (`App` struct)

```go
HasProfile() (bool, error)
CreateProfile(input) (Profile, error)     // one txn: profile row + optional first weight_log
GetProfile() (*Profile, error)
UpdateProfile(input) (Profile, error)
AddWeightLog(weightKg float64, loggedAt string) (WeightLog, error)
ListWeightLogs() ([]WeightLog, error)

GetActiveFast() (*Fast, error)            // drives restart rehydration
StartFast(plannedDurationHours float64) (Fast, error)
CompleteFast(id string, mood int, note string) (Fast, error)
AbandonFast(id string, mood int, note string) (Fast, error)
ListFasts(limit, offset int) ([]Fast, error)

GetStats() (Stats, error)                 // total, completionRate%, longest, average, currentStreak
```
Mood trend for Reports reuses the `ListFasts` payload rather than a dedicated call, keeping the IPC surface small.

**Stats logic** (`internal/db/stats.go`): completion rate and longest/average duration are simple SQL aggregates over terminal fasts (average counts `completed` only, so abandoned attempts don't skew "typical successful fast length"). Streak is computed in Go by grouping completed fasts by local calendar date and walking backward from today (skipping today if it has no completed fast yet, since an active in-progress fast shouldn't appear to break yesterday's streak).

## Frontend structure (Vite + vanilla TS)

3 screens behind a ~50-line hash router (`#/profile #/fasting #/reports`, plus `#/onboarding` on first run), a tiny pub/sub `state.ts` (caches profile + active-fast status for the nav bar), and a typed `api.ts` wrapper over the generated Wails bindings. No framework, no date library, no state-management library.

- `screens/onboarding.ts` — username + DOB (required), weight + height with unit toggles (optional) → single `CreateProfile` call.
- `screens/fasting.ts` — idle: preset tiles (24/36/48/72/120h + custom) → `StartFast`. Active: ring timer + current fasting-stage callout + disclaimer + "End Fast" → mood/note modal → `CompleteFast` (goal reached) or `AbandonFast` (ended early).
- `screens/reports.ts` — history list from `ListFasts`, stat cards from `GetStats`, weight/mood trend mini-charts.
- `components/ring-timer.ts` — SVG ring, `circumference = 2πr` computed once, one `strokeDashoffset` write/second, starts full (`dashoffset=0`) and drains toward `dashoffset=circumference` — matches "starts full, decrements, shows remaining."

### Fasting stages (static educational content, with disclaimer)
0–4h Fed State → 4–16h Post-Absorptive → 12–18h Glycogen Depletion → 18–24h Rising Ketosis → 24–48h Ketosis Established → 48–72h Deepening Autophagy → 72h+ Extended Fasting/Immune Renewal. Persistent disclaimer on the Fasting screen: *"Educational information only, based on general fasting research — not medical advice. Consult a healthcare provider before extended fasting, especially beyond 48 hours."*

### Color palette — lively monochromatic (CSS custom properties)
Single `--hue` variable drives everything; only lightness/saturation vary, so swapping the app's color is a one-line change:
```css
--hue: 190;                                    /* placeholder teal-cyan */
--color-bg: hsl(var(--hue) 20% 8%);            --color-surface: hsl(var(--hue) 18% 12%);
--color-text: hsl(var(--hue) 15% 92%);         --color-text-muted: hsl(var(--hue) 10% 65%);
--color-accent: hsl(var(--hue) 85% 55%);       /* ring progress, primary buttons, active nav */
--color-accent-strong: hsl(var(--hue) 90% 65%);
```
Dark theme by default; fasting outcome (completed vs. abandoned) is distinguished by lightness/label, not a hue shift, to stay strictly monochromatic. `--hue` is a placeholder — easy to swap for the user's preferred color later.

## Build order

1. Environment setup + `wails doctor` green.
2. Scaffold (`wails init -t vanilla-ts`), merge into repo, confirm default template runs, first commit.
3. Data layer: `modernc.org/sqlite` + schema/migrations (via `PRAGMA user_version`) + CRUD + Go unit tests (`go test ./...` green before touching UI).
4. Go↔JS bindings wired in `main.go`; confirm generated `wailsjs` types.
5. App shell: palette tokens, router, nav, empty placeholder screens (prove routing end-to-end).
6. Onboarding + Profile screens.
7. Fasting core as plain countdown text first (validate `StartFast`/rehydration/`CompleteFast`/`AbandonFast` logic) before adding the ring visual.
8. Ring timer + fasting-stage content.
9. Reports: history, stats cards, trend charts.
10. Polish: validation (DOB not future, custom duration 1–240h bounds), error states, app icon.
11. `wails build` verification pass.

### Critical files
- `internal/db/db.go`, `internal/db/fasts.go` — connection setup, migrations, core CRUD
- `app.go` — Go↔JS binding surface
- `frontend/src/components/ring-timer.ts` — circular countdown
- `frontend/src/screens/fasting.ts` — main fasting flow
- `frontend/src/style.css` — monochromatic palette tokens

## Verification

- `wails dev` for interactive hot-reload testing throughout.
- `go test ./...` against `:memory:` SQLite: CRUD correctness, the one-active-fast constraint, duration math, and stats edge cases (streak gaps, abandoned fasts excluded from streak/average).
- Manual click-through: fresh onboarding → start/rehydrate-after-restart/complete/abandon a fast → Reports stats match hand-computed expectations → weight logs update the trend chart with correct unit conversion → confirm only one timer interval is ever live (no stacking on repeated navigation) → `wails build` produces a standalone binary whose DB lands under `os.UserConfigDir()`, not cwd.
