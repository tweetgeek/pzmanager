package launcher

import (
	"context"
	"fmt"
	"log"
	"sort"

	"pzm/internal/config"
	"pzm/internal/domain"
	"pzm/internal/storage"
)

// GamePlatform — wąski podzbiór platform.Platform potrzebny launcherowi.
// Definiowany lokalnie, by łatwiej fake'ować w testach.
type GamePlatform interface {
	LaunchGame(inst domain.GameInstance, steamCommand []string) error
	DetectGame() (domain.GameInstance, error)
}

type Launcher struct {
	scanner  storage.SaveScanner
	store    storage.BackupStore
	platform GamePlatform
	cfg      *config.Settings
}

func New(sc storage.SaveScanner, st storage.BackupStore, p GamePlatform, cfg *config.Settings) *Launcher {
	return &Launcher{scanner: sc, store: st, platform: p, cfg: cfg}
}

// AutoLaunch wykonuje sekwencję: detekcja zmian → (opcjonalny) backup →
// prune → uruchomienie gry. Błędy backupu logowane, ale gra startuje
// niezależnie (priorytet: nie blokować rozgrywki).
func (l *Launcher) AutoLaunch(ctx context.Context, steamCmd []string, ui ProgressReporter) error {
	inst, err := l.platform.DetectGame()
	if err != nil {
		// detect może zawieść gdy user wskazał ścieżkę ręcznie — buduj instance z config
		inst = domain.GameInstance{InstallPath: l.cfg.GameInstallPath, SavesPath: l.cfg.SavesPath}
	}

	skipCh := ui.OnSkip()
	skipped := false

	if l.cfg.AutoBackup {
		ui.Status("Looking for active save...")
		ui.Progress(0.05)

		save, found, scanErr := l.activeSave()
		if scanErr != nil {
			log.Printf("scan saves: %v", scanErr)
		} else if found {
			// sprawdź czy user kliknął skip
			select {
			case <-skipCh:
				skipped = true
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if !skipped {
				changed, hashErr := l.saveChanged(save)
				if hashErr != nil {
					log.Printf("hash check: %v (will create backup as fallback)", hashErr)
				}
				if changed {
					ui.Status("Creating backup...")
					ui.Progress(0.2)

					format, _ := domain.ParseBackupFormat(l.cfg.DefaultFormat)
					if format == 0 {
						format = domain.BackupFormatZip
					}
					_, createErr := l.store.Create(save, storage.CreateOpts{
						Format: format,
						Kind:   domain.BackupKindAuto,
					})
					if createErr != nil {
						log.Printf("create backup: %v (continuing — game will still launch)", createErr)
					} else {
						ui.Progress(0.7)
						ui.Status("Pruning old auto-backups...")
						if _, pruneErr := l.store.PruneAuto(save.ID, l.cfg.AutoBackupRetain); pruneErr != nil {
							log.Printf("prune: %v", pruneErr)
						}
					}
				}
			}
		}
	}

	ui.Status("Launching game...")
	ui.Progress(0.95)
	if err := l.platform.LaunchGame(inst, steamCmd); err != nil {
		return fmt.Errorf("launch game: %w", err)
	}
	ui.Progress(1.0)
	return nil
}

// activeSave wybiera save z najnowszym LastPlayed. Zwraca (Save, true) gdy znaleziono.
func (l *Launcher) activeSave() (domain.Save, bool, error) {
	saves, err := l.scanner.Scan(l.cfg.SavesPath)
	if err != nil {
		return domain.Save{}, false, err
	}
	if len(saves) == 0 {
		return domain.Save{}, false, nil
	}
	sort.Slice(saves, func(i, j int) bool { return saves[i].LastPlayed.After(saves[j].LastPlayed) })
	return saves[0], true, nil
}

// saveChanged porównuje aktualny hash z hashem ostatniego auto-backupu.
// Gdy nie wiemy (błąd / brak poprzedniego) — zwracamy true (bezpieczna opcja: zrób backup).
func (l *Launcher) saveChanged(save domain.Save) (bool, error) {
	fresh, err := l.scanner.Hash(save)
	if err != nil {
		return true, err
	}
	backups, err := l.store.List(save.ID)
	if err != nil {
		return true, err
	}
	var lastAuto *domain.Backup
	for i := range backups {
		if backups[i].Kind == domain.BackupKindAuto {
			if lastAuto == nil || backups[i].CreatedAt.After(lastAuto.CreatedAt) {
				lastAuto = &backups[i]
			}
		}
	}
	if lastAuto == nil {
		return true, nil
	}
	return lastAuto.SaveHash != fresh, nil
}
