# Project Zomboid Manager

Cross-platform desktop launcher and save backup manager for Project Zomboid (Windows, Linux, macOS).

Built with Go and [Fyne v2](https://fyne.io/).

## Features

- List Project Zomboid saves with metadata (game version, last played, size, backup count)
- Manual backups (ZIP or folder copy) with optional labels
- Auto-backup before launching the game (only when save changed; configurable retention)
- Atomic restore with optional safety backup of current save
- Three launcher integration modes: from app, desktop shortcut, Steam Launch Options
- Dedicated progress window when launching from Steam (with skip option)

## Build

Requires Go 1.26+ and a C compiler (Fyne uses CGO).

```bash
go build -ldflags "-X main.Version=$(git describe --tags --always 2>/dev/null || echo dev)" -o bin/pz-manager ./cmd/manager
```

For platform-specific packaging (icons, .app bundles, .exe metadata):

```bash
go install fyne.io/tools/cmd/fyne@latest
fyne package -os darwin   # or linux / windows
```

Cross-compiling Fyne apps from macOS/Linux to Windows etc. requires platform-specific C toolchains; the easiest workflow is to build natively on each OS or use CI runners.

## Run

```bash
./bin/pz-manager                      # GUI mode
./bin/pz-manager --auto-launch        # auto-backup + launch game (CLI)
./bin/pz-manager --version            # print version
```

For Steam integration, in Steam → Properties → Launch Options:

```
"<absolute-path-to-pz-manager>" --auto-launch -- %command%
```

## Test

```bash
go test ./...
```

## Releases

The project uses GitHub Actions for distribution. See `.github/workflows/`:

- **`ci.yml`** — runs on every push / pull request to `main`. Verifies `go vet`, `go test`, and a Linux build. Catches regressions before tagging.
- **`release.yml`** — runs on tags matching `v*`. Builds binaries for Linux amd64, macOS arm64, and Windows amd64 on native runners (Fyne requires CGO so cross-compilation is impractical), then publishes a GitHub Release with the binaries attached.

### How to cut a release

```bash
# 1. Tag locally (annotated tag with notes)
git tag -a v0.2.0 -m "Description of what changed"

# 2. Push the tag
git push origin v0.2.0
```

The `release.yml` workflow picks up the push, runs the matrix build, and creates the GitHub Release. Pre-releases (tags containing `-`, e.g. `v0.2.0-rc1`) are marked as such automatically.

Released artifacts:

- `pz-manager-linux-amd64`
- `pz-manager-darwin-arm64`
- `pz-manager-windows-amd64.exe`

The version printed by `pz-manager --version` is injected at build time from the tag name via `-ldflags "-X main.Version=$GITHUB_REF_NAME"`.

## Project structure

```
.github/workflows/          # CI + release pipelines
cmd/manager/                # entrypoint, CLI args, DI
internal/
  domain/                   # core types: Save, Backup, GameInstance (no I/O)
  storage/                  # filesystem ops: scan, ZIP, restore, retention
  platform/                 # OS-specific paths and process launching
  launcher/                 # auto-launch flow orchestration
  config/                   # settings.json (atomic write + migration)
  ui/                       # Fyne views, dialogs, progress window
docs/superpowers/
  specs/                    # design spec
  plans/                    # implementation plan
```

## Documentation

- [Design spec](docs/superpowers/specs/2026-05-06-pz-manager-design.md)
- [Implementation plan](docs/superpowers/plans/2026-05-06-pz-manager.md)

## License

TBD.
