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
- Windows code signing (Authenticode) — `release.yml` produces an unsigned `.exe`. SmartScreen warns on first run; user clicks "More info → Run anyway". Real signing needs a code-signing cert (DigiCert / SSL.com / etc., ~$200+/yr) + `signtool.exe`. Out of scope.
- `.dmg` for macOS instead of `.app.zip` — would need `create-dmg` or similar. ZIP works fine.
- macOS Intel (amd64) builds — only ARM64 in release matrix
- I18n / Polish UI labels — UI text is EN-only
- Mock-based UI testing — Fyne tests use `test.NewApp()` headless driver, smoke-only
- Auto-update / telemetry / tray icon

## Branding / icon

- Master source: `assets/icon-source.jpg` (585×585 JPEG, original artwork — backpack + save disk + folder + gear, "PZ MANAGER" text).
- Production asset: `assets/icon.png` (512×512 PNG, generated from JPG via `sips -s format png -Z 512`).
- Two consumers, both with their own copy of the PNG (manual sync — re-copy from `assets/icon.png` if the source changes):
  - `internal/ui/icon.png` — embedded into Fyne binary via `//go:embed` in `internal/ui/app.go` as `ui.AppIcon` (`fyne.StaticResource`). Set on the Fyne app and on every window (`fyneApp.SetIcon`, `window.SetIcon`).
  - `web/static/icon.png` — served as `/icon.png` by Hugo. Used in `web/layouts/index.html` as favicon, apple-touch-icon, and OpenGraph image.
- To regenerate the PNGs after updating `assets/icon-source.jpg`:
  ```bash
  sips -s format png -Z 512 assets/icon-source.jpg --out assets/icon.png
  cp assets/icon.png internal/ui/icon.png
  cp assets/icon.png web/static/icon.png
  ```
- File-system icons (Finder thumbnail, .exe resource icon, .icns inside .app bundle) are set via `fyne package` in `release.yml`. The CI converts `internal/ui/icon.png` → `.icns` (macOS) and `.ico` (Windows) automatically. See `FyneApp.toml` at repo root for the metadata fed to `fyne package` (Name, ID, Icon, Version) — `Version` and `Icon` (path) are overridden by CI flags; `Name` and `ID` come from the TOML.

## macOS code signing + notarization

The release workflow produces a signed and notarized `.app` for macOS so users don't see Gatekeeper warnings on first launch.

- **Apple Developer Program** ($99/yr) is required. Provides the Developer ID Application certificate + access to the notary service.
- **Certificate name** is auto-detected from the imported keychain at runtime (`security find-identity -v -p codesigning | grep "Developer ID Application"`). No need to hardcode the team name in the workflow — only the Team ID is implicit through the cert.
- **Required GitHub repo secrets:**
  - `APPLE_CERT_P12_BASE64` — base64-encoded `.p12` export of the Developer ID Application cert (run `base64 -i Certificates.p12` on the source machine)
  - `APPLE_CERT_PASSWORD` — password set when exporting the `.p12`
  - `APPLE_API_KEY_ID` — App Store Connect API Key ID
  - `APPLE_API_ISSUER_ID` — App Store Connect API Issuer ID
  - `APPLE_API_KEY_P8_BASE64` — base64-encoded `.p8` private key file (downloaded once from App Store Connect; can't be re-downloaded — store it carefully)
  - `APPLE_TEAM_ID` — currently unused by the workflow but kept as a reference; the cert's Team ID is sufficient
- **Flow** (in `release.yml`, darwin matrix only):
  1. `Import macOS signing certificate` — decodes p12, creates ephemeral keychain, imports cert, exports `MAC_CERT_NAME` env var
  2. `Package binary with fyne package` — passes `--certificate "$MAC_CERT_NAME"` so codesign runs with hardened runtime + secure timestamp (both required for notarization)
  3. `Notarize and staple (macOS)` — submits zipped `.app` to `xcrun notarytool` (waits ~1-5 min), then `xcrun stapler staple` attaches the ticket so it works offline
- **Notary log on failure** — if `notarytool submit --wait` exits non-zero, currently you'd need to follow up with `xcrun notarytool log <submission-id>` to see what Apple flagged. Could be wired into the workflow on failure (`if: failure()` step) — not done yet.
- **Renewing the cert** — Developer ID Application certs expire after 5 years. Re-export, base64, update the GH secret. The `notarytool` API key has its own (configurable) lifetime.

## Documentation

- Design spec: `docs/superpowers/specs/2026-05-06-pz-manager-design.md`
- Implementation plan (24 tasks, all completed): `docs/superpowers/plans/2026-05-06-pz-manager.md`
- Original website plan (superseded): `docs/superpowers/plans/2026-05-07-public-website.md`
- Hugo + Cloudflare migration plan: `docs/superpowers/plans/2026-05-07-hugo-cloudflare-migration.md`

## Public website (Hugo + Cloudflare Pages)

- Source: `web/` — Hugo project with own custom layouts (no external theme).
- Deploy: `.github/workflows/pages.yml` listens for `release: published/edited/released`, `workflow_dispatch`, and `push` to `main` touching `web/**`. Builds with Hugo (`peaceiris/actions-hugo@v3`, pinned to `0.160.1` extended), deploys to Cloudflare Pages via `cloudflare/wrangler-action@v3` direct upload.
- Cloudflare project name in workflow: `pz-manager` — must match the project created manually in the Cloudflare dashboard. Renaming requires updating the workflow.
- Required GitHub repo secrets: `CLOUDFLARE_API_TOKEN`, `CLOUDFLARE_ACCOUNT_ID`. Without these the workflow fails fast at the Wrangler step.
- Data contract (`web/data/releases.json`): assembled at build time by `gh api --paginate` + `jq` from the Releases API. Shape: `{ generated_at, repository, latest, releases: [{ tag, name, published_at, prerelease, body_md, assets: { "linux-amd64"|"darwin-arm64"|"windows-amd64": { name, url, size } } }] }`. Sorted newest-first; drafts excluded; `latest` = newest non-prerelease falling back to newest overall. `web/data/releases.json` is gitignored (CI generates it); `web/data/releases.example.json` is committed for local dev.
- Hugo template (`web/layouts/index.html`) renders downloads table + changelog **statically** at build time — no client-side fetch, no runtime markdown render. Uses `markdownify` (Goldmark) for release notes, `partial "humanizeSize.html"` for byte formatting, and `range $key, $cfg := site.Params.targets` to iterate the three OS targets defined in `web/hugo.toml`.
- **Hugo data API:** the template uses `hugo.Data.releases` (not `.Site.Data.releases` — that was deprecated in Hugo 0.156+). `site.Params` is still fine.
- JS only enhances: `web/static/js/os-detect.js` reads `navigator.userAgent`, sets `body[data-target]`, adds `recommended` class on the matching downloads row, and unhides the matching primary CTA variant (hiding the fallback). Page is fully usable with JS off.
- **Asset filenames in `release.yml` are load-bearing.** The `jq` filter looks up assets by exact name: `pz-manager-linux-amd64`, `pz-manager-darwin-arm64`, `pz-manager-windows-amd64.exe`. Renaming in one workflow without updating the other will silently break the data contract.
- **`jq` gotcha:** `asset_for` uses `// null` (not `// empty`); with `empty` the entire enclosing release silently disappears from the array because `empty` propagates through object construction.
- Local dev: `cp web/data/releases.example.json web/data/releases.json && hugo --source web serve` → http://localhost:1313/. Tests at `/js/tests.html`.
- Hugo install: `brew install hugo` (macOS) or download from gohugoio releases. Must be the **extended** edition.
- **Privacy:** Cloudflare Pages supports private repos on the free plan. The site itself is public regardless of repo visibility — it exposes only release metadata + asset download URLs, never source.

## Renaming history

Project was originally **"Project Zomboid Save Manager"** (module `pz_sm`, binary `pz-save-manager`). Renamed to **"Project Zomboid Manager"** (module `pzm`, binary `pz-manager`) along with all default paths (`PZSaveManager` → `PZManager`, `pz-save-manager` → `pz-manager`). The on-disk project directory `ProjectZomboidSaveManager/` was intentionally kept — renaming it requires shell-level move + Go workspace path update; do only if user explicitly asks.
