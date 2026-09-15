# Faster — Fasting Tracker

A private, offline-first desktop app for tracking fasting habits. Built with
[Wails v2](https://wails.io) (Go backend + vanilla TypeScript frontend) and
SQLite. No accounts, no network calls, no telemetry — everything lives in a
local database on your machine.

See [`docs/PLAN.md`](docs/PLAN.md) for the full architecture writeup
(schema, Go↔JS binding surface, palette system, etc.).

## Installing a release

Grab the latest build from the [Releases page](https://github.com/BlasioMitch/faster/releases)
instead of building from source, unless you're developing the app.

### Windows

Download `faster_<version>_windows_amd64.zip`, extract it anywhere, and
run `install.ps1` from that folder (right-click it → **Run with
PowerShell**; see the included `README.txt` if Windows blocks it with a
script-execution-policy message). It adds both a **Start Menu** entry and
a **Desktop** shortcut for the current user — no admin rights needed.

Prefer not to install anything? Just run `faster.exe` directly from the
extracted folder — the shortcuts are optional.

To remove them later, run `uninstall.ps1` the same way (your data isn't
touched). A traditional single-file installer (`.exe`, with an
uninstaller registered in Windows' Add/Remove Programs) may be added in a
future release.

### Linux

Download `faster_<version>_linux_amd64.tar.gz`, extract it, and run the
installer it contains:

```bash
tar -xzf faster_*_linux_amd64.tar.gz
cd faster_*_linux_amd64   # or wherever you extracted it
./install.sh
```

This is a **user-level install — no `sudo` needed**. It copies the binary
to `~/.local/bin/faster`, registers an icon, and adds an entry to your
application menu (the Linux equivalent of the Start Menu; it may take a
moment, or a re-login, to show up depending on your desktop environment).
The script prints a one-line command at the end for pinning a launcher
icon to your Desktop too, since that step needs a `gio` trust flag on
GNOME-based desktops.

To remove it later, run `./uninstall.sh` from the same extracted folder
(your data in `~/.config/faster` is left untouched).

Prefer not to install anything? Just extract the tarball and run `./faster`
directly — it works standalone, no installation required.

## Building from source

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
