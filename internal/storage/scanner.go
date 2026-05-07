package storage

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"pzm/internal/domain"
)

// SaveScanner reads the list of saves from the game's saves directory.
type SaveScanner interface {
	Scan(savesRoot string) ([]domain.Save, error)
	Hash(save domain.Save) (string, error)
}

type fsSaveScanner struct{}

func NewFsSaveScanner() SaveScanner { return &fsSaveScanner{} }

// Scan iterates over directories at <savesRoot>/<GameMode>/<SaveName>/.
func (s *fsSaveScanner) Scan(savesRoot string) ([]domain.Save, error) {
	info, err := os.Stat(savesRoot)
	if err != nil {
		return nil, fmt.Errorf("saves root: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("saves root %q is not a directory", savesRoot)
	}

	var saves []domain.Save

	gameModes, err := os.ReadDir(savesRoot)
	if err != nil {
		return nil, fmt.Errorf("read saves root: %w", err)
	}

	for _, gm := range gameModes {
		if !gm.IsDir() {
			continue
		}
		gameModePath := filepath.Join(savesRoot, gm.Name())
		saveDirs, err := os.ReadDir(gameModePath)
		if err != nil {
			log.Printf("scan: failed to read game mode dir %q: %v", gameModePath, err)
			continue
		}
		for _, sd := range saveDirs {
			if !sd.IsDir() {
				continue
			}
			savePath := filepath.Join(gameModePath, sd.Name())
			saves = append(saves, s.buildSave(gm.Name(), sd.Name(), savePath))
		}
	}
	return saves, nil
}

func (s *fsSaveScanner) Hash(save domain.Save) (string, error) {
	return HashSaveContents(save.Path)
}

func (s *fsSaveScanner) buildSave(gameMode, name, path string) domain.Save {
	size, lastMod, err := dirSizeAndMtime(path)
	if err != nil {
		log.Printf("scan: failed to compute size/mtime for %q: %v", path, err)
		// size and lastMod will be zero values — acceptable
	}
	return domain.Save{
		ID:         HashSavePath(path),
		Name:       name,
		GameMode:   gameMode,
		Path:       path,
		LastPlayed: lastMod,
		SizeBytes:  size,
	}
}

func dirSizeAndMtime(root string) (int64, time.Time, error) {
	var total int64
	var latest time.Time
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		total += info.Size()
		if info.ModTime().After(latest) {
			latest = info.ModTime()
		}
		return nil
	})
	return total, latest, err
}
