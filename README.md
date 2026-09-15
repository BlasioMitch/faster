# Faster — Fasting Tracker

A private, offline-first desktop app for tracking fasting habits. Built with
[Wails v2](https://wails.io) (Go backend + vanilla TypeScript frontend) and
SQLite. No accounts, no network calls, no telemetry — everything lives in a
local database on your machine.

See [`docs/PLAN.md`](docs/PLAN.md) for the full architecture writeup
(schema, Go↔JS binding surface, palette system, etc.).

## Requirements

- Go 1.22+
- Node.js 20.19+ or 22.12+ (Vite 7 requirement; slightly older Node mostly
  still works but prints a warning)
- The [Wails CLI](https://wails.io/docs/gettingstarted/installation):
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```
- Linux only: WebKitGTK + GTK3 development headers.
  ```bash
  sudo apt install -y libwebkit2gtk-4.1-dev build-essential libgtk-3-dev pkg-config
  ```

### Note for Ubuntu 24.04+ / any distro shipping only WebKitGTK 4.1

Ubuntu 24.04 dropped the `webkit2gtk-4.0` package in favor of `4.1`, and
Wails v2's default build looks for `4.0`. `wails.json` already sets
`"build:tags": "webkit2_41"` so plain `wails dev`/`wails build` pick the
right WebKitGTK version automatically on such systems — no flag needed.

(`wails doctor` may still report `libwebkit` as "Not Found" here even
though the 4.1 dev package is installed — that's a cosmetic detection gap
in that version of the doctor command, not a real problem;
`pkg-config --exists webkit2gtk-4.1` is the source of truth.)

## Development

```bash
wails dev
```

Hot-reloads both the Go backend and the TypeScript frontend.

## Production build

```bash
wails build
```

Produces a standalone binary at `build/bin/faster`. The app's SQLite
database lives under the OS's per-user config directory (e.g.
`~/.config/faster/faster.db` on Linux), never next to the binary.

## Tests

```bash
go test ./...
```

Covers the SQLite data-access layer: profile/weight-log CRUD, the
one-active-fast-at-a-time constraint, fast completion/abandonment duration
math, and streak/stats edge cases.

## Project layout

```
internal/db/       SQLite schema, migrations, CRUD, stats
internal/models/    Shared Go structs (JSON contract with the frontend)
app.go              Go↔JS binding surface (the App struct Wails binds)
main.go             Entry point: opens the DB, launches the Wails window
frontend/src/
  screens/          profile.ts, fasting.ts, reports.ts, onboarding.ts
  components/       nav.ts, ring-timer.ts, modal.ts, mini-chart.ts
  data/             static presets, fasting-stage content, mood scale
  router.ts, state.ts, api.ts   app shell plumbing (no framework)
```
