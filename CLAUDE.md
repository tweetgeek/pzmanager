# Project Zomboid Manager — Agent Notes

Cross-platform desktop app (Win/Linux/macOS) that launches Project Zomboid and manages save backups. Built with Go + Fyne v2.

## Workflow rules (READ FIRST)

- **DO NOT run `git commit`.** The user commits manually. Stop after the working tree changes are ready and summarize what was modified — do not stage with intent to commit on their behalf.
- **DO NOT run `git push`** (or `git push --force`, tag pushes, branch pushes — anything that touches a remote). The user handles all remote interactions.
- `git status`, `git diff`, `git log`, `git add` (when explicitly asked to stage), and other read-only / local-only commands are fine.
- If a task seems to require a commit (e.g. "tag a release"), prepare the changes and tell the user what command to run — don't run it.

## Project identity

- **Display name:** Project Zomboid Manager
- **Binary:** `pz-manager`
- **Go module:** `pzm`
- **Working directory on disk:** `/Users/sdawidziak/_TweetGeek/_Projects/ProjectZomboidSaveManager` (legacy folder name; not renamed)
- **License:** TBD

## Tech stack

- Go 1.26+ (see `go.mod`)
- Fyne v2.7.3 (GUI; uses CGO — cross-compile impractical, build natively per OS)
- Std lib only beyond Fyne (`archive/zip`, `crypto/sha256`, `encoding/json`, `os/exec`)
- Tests: `testing` + `t.TempDir()`; UI: `fyne.io/fyne/v2/test` headless driver
- UI language: **EN** (chosen for larger PZ community)

## Layered architecture

```
cmd/manager/                    entrypoint, CLI args, dependency injection
internal/
├── domain/                     pure types: Save, Backup, GameInstance (no I/O)
├── storage/                    FS ops: scan, hash, ZIP, restore (atomic), retention
├── platform/                   OS-specific paths, DetectGame, LaunchGame, shortcuts
│                               (per-OS files with build tags; New() lives in each)
├── launcher/                   AutoLaunch orchestration + ProgressReporter interface
├── config/                     Settings (JSON, atomic save) + Defaults()
└── ui/                         Fyne views: nav, saves, detail, dialogs, settings,
                                progress window, about, first-run wizard
```

Each layer behind an interface — easy to fake in tests. UI receives a `Services` struct via constructor (no globals, no singletons).

## Commands

```bash
# Build (native OS)
go build -ldflags "-X main.Version=$(git describe --tags --always 2>/dev/null || echo dev)" \
  -o bin/pz-manager ./cmd/manager

# Run
./bin/pz-manager                      # GUI
./bin/pz-manager --auto-launch        # CLI: backup + launch then exit
./bin/pz-manager --auto-launch -- %command%   # Steam Launch Options form
./bin/pz-manager --version

# Quality
go test ./...
go vet ./...

# Release: tag triggers full matrix build via GitHub Actions
git tag -a v0.x.0 -m "Notes"
git push origin v0.x.0
```

Pre-release tags: anything containing `-` (e.g. `v0.2.0-rc1`) is auto-marked prerelease in `release.yml`.

## Default paths (per OS)

| OS | Saves | Game install (Steam) | Config | Backups |
|---|---|---|---|---|
| Windows | `%USERPROFILE%\Zomboid\Saves` | `Program Files (x86)\Steam\steamapps\common\ProjectZomboid` | `%APPDATA%\PZManager` | `%USERPROFILE%\Documents\PZManager\Backups` |
| Linux | `~/Zomboid/Saves` | `~/.steam/steam/steamapps/common/ProjectZomboid` | `~/.config/pz-manager` | `~/PZManager/Backups` |
| macOS | `~/Zomboid/Saves` | `~/Library/Application Support/Steam/steamapps/common/ProjectZomboid` | `~/Library/Application Support/PZManager` | `~/Documents/PZManager/Backups` |

Steam AppID for PZ: `108600`.

## Product decisions (locked)

| # | Topic | Decision |
|---|---|---|
| 1 | Backup format | ZIP (default) + 1:1 folder copy "fast mode"; chosen per-backup or in settings |
| 2 | Auto-backup | Only when save changed since last auto-backup (SHA256 of {relpath, size, mtime}) + retention of N latest auto-backups; manual backups never deleted automatically; user toggle on/off |
| 3 | Launcher integration | Wizard offers 3 modes (combinable): (A) launch from app, (B) desktop shortcut → app, (C) Steam Launch Options string copied to clipboard |
| 4 | MVP scope | Core only. **Excluded:** game profiles, mod management, server saves, export/import backup ZIP, cloud sync, auto-update, tray icon |
| 5 | UI layout | Sidebar nav (▶ Launch / 📁 Saves / ⚙ Settings / ℹ About) + drill-down from saves list to detail |
| 6 | Restore safety | Confirm dialog with "Back up current save before restoring" checkbox **checked by default** |
| 7 | Auto-launch UX | Small ProgressWindow with status + bar + "Skip backup, launch now" button |

## Key invariants and conventions

- **Domain has zero I/O.** Only stdlib `time`, `fmt`, `strings`, `encoding/json` allowed.
- **`Save.ID`** = `HashSavePath(absolute path)` — stable across renames within same path; resilient to same-name saves in different game modes.
- **`Backup.SaveHash`** = `HashSaveContents(saveDir)` over file metadata — used by launcher to decide "did the save change since last auto-backup?".
- **`BackupKind`/`BackupFormat`** serialize to JSON as **strings** (`"auto"`, `"manual"`, `"zip"`, `"folder"`) via custom `MarshalJSON`/`UnmarshalJSON`. Backup metadata files must remain human-readable for manual inspection.
- **Atomic writes everywhere** — `Settings.Save`, backup metadata, restore swap. Pattern: write to `.tmp` → `os.Rename`. On rename failure, clean up `.tmp`.
- **Restore is atomic-swap** — extract/copy to `<target>.restore-tmp`, move existing target to `<target>.restore-old`, rename tmp → target, then remove old. On second-rename failure, attempt rollback if `.restore-old` still exists.
- **Zip-slip protected** — `unzipTo` validates each entry's resolved path is inside `dst`.
- **Path-traversal protected** — `Delete` and `PruneAuto` call `assertWithinRoot(b.Path)` before any destructive op (defends against malformed metadata JSON).
- **`PruneAuto`** never touches `Kind=Manual`. On partial-delete error it returns `(deletedSoFar, err)` — caller decides retry vs. accept.
- **`BackupStore` is not internally locked.** Caller (UI / launcher) must serialize per-backup ops. Acceptable for single-user desktop app.
- **`crypto/rand.Read` errors panic** in `newBackupID()` — extremely rare and indicates a broken entropy source; safer to fail loud than collide IDs.
- **Backup failures must NOT block game launch.** `launcher.AutoLaunch` logs and proceeds to `LaunchGame` even when create/prune fails. Priority: never block player from playing.
- **UI updates from goroutines** — always wrap in `fyne.Do(...)`. Fyne crashes if `Refresh()` is called off the main thread.

## Per-OS implementation notes

- Each `internal/platform/<os>.go` has its own `func New() Platform` with build tag — no `runtime.GOOS` switch needed, no cross-OS stub constructors.
- **Linux launch:** `xdg-open steam://run/108600` (Steam fallback) or direct `<install>/projectzomboid`. Shortcut: `.desktop` file in `~/Desktop` or `~/.local/share/applications`.
- **macOS launch:** `open steam://run/108600` (Steam fallback) or `open -a "Project Zomboid"`. Shortcut: `.command` script on Desktop.
- **Windows launch:** `cmd /c start steam://run/108600` (Steam) or `<install>\ProjectZomboid64.exe`. Shortcut: `.lnk` via PowerShell + WScript.Shell COM (no external deps).

## CLI argument convention

```
pz-manager                              # GUI
pz-manager --version                    # print and exit
pz-manager --verbose                    # INFO-level logging (default WARN+)
pz-manager --auto-launch                # CLI: do backup, launch game, exit
pz-manager --auto-launch -- <args...>   # Steam Launch Options: args after `--`
                                        # are passed verbatim to LaunchGame as
                                        # steamCmd (i.e. %command% from Steam)
```

## CI/CD

`.github/workflows/`:

- **`ci.yml`** — push/PR to `main`: `go vet` + `go test` + Linux build (with Fyne deps `libgl1-mesa-dev xorg-dev libxxf86vm-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev`).
- **`release.yml`** — push of tag `v*`: matrix build on `ubuntu-latest` / `macos-14` / `windows-latest`. Publishes a GitHub Release via `softprops/action-gh-release@v2` with assets:
  - `pz-manager-linux-amd64`
  - `pz-manager-darwin-arm64` (Apple Silicon only; Intel macOS not in matrix)
  - `pz-manager-windows-amd64.exe`
  - Binaries built with `-trimpath -ldflags "-s -w -X main.Version=$GITHUB_REF_NAME"`.

**Repo is not yet on GitHub** as of last touch — workflows activate after first push to a GitHub remote.

## Things deliberately NOT done (don't re-introduce without asking)

- Code signing / notarization (Apple, Windows) — needs cert secrets, out of MVP scope
- `fyne package` (proper `.app`/`.exe` bundles with embedded icons) — needs `FyneApp.toml` and `icon.png`. Workflow uses plain `go build` instead. There's an untracked `pzmod.jpg` in repo root that may become the icon source — confirm with user before integrating.
- macOS Intel (amd64) builds — only ARM64 in release matrix
- I18n / Polish UI labels — UI text is EN-only
- Mock-based UI testing — Fyne tests use `test.NewApp()` headless driver, smoke-only
- Auto-update / telemetry / tray icon

## Documentation

- Design spec: `docs/superpowers/specs/2026-05-06-pz-manager-design.md`
- Implementation plan (24 tasks, all completed): `docs/superpowers/plans/2026-05-06-pz-manager.md`

## Renaming history

Project was originally **"Project Zomboid Save Manager"** (module `pz_sm`, binary `pz-save-manager`). Renamed to **"Project Zomboid Manager"** (module `pzm`, binary `pz-manager`) along with all default paths (`PZSaveManager` → `PZManager`, `pz-save-manager` → `pz-manager`). The on-disk project directory `ProjectZomboidSaveManager/` was intentionally kept — renaming it requires shell-level move + Go workspace path update; do only if user explicitly asks.
