package storage

import (
	"os"
	"path/filepath"
	"sort"

	"pzm/internal/domain"
)

// PruneAuto kasuje najstarsze backupy Kind=Auto poza najnowszymi `keep`.
// Backupy Kind=Manual NIE SĄ tykane. Zwraca liczbę usuniętych.
func (s *fsBackupStore) PruneAuto(saveID string, keep int) (int, error) {
	if keep < 0 {
		keep = 0
	}
	backups, err := s.List(saveID)
	if err != nil {
		return 0, err
	}
	var autos []domain.Backup
	for _, b := range backups {
		if b.Kind == domain.BackupKindAuto {
			autos = append(autos, b)
		}
	}
	sort.Slice(autos, func(i, j int) bool { return autos[i].CreatedAt.After(autos[j].CreatedAt) })
	if len(autos) <= keep {
		return 0, nil
	}
	toDelete := autos[keep:]
	saveDir := filepath.Join(s.root, saveID)
	deleted := 0
	for _, b := range toDelete {
		if err := s.assertWithinRoot(b.Path); err != nil {
			return deleted, err
		}
		if err := os.RemoveAll(b.Path); err != nil && !os.IsNotExist(err) {
			return deleted, err
		}
		metaPath := filepath.Join(saveDir, b.ID+".json")
		if err := os.Remove(metaPath); err != nil && !os.IsNotExist(err) {
			return deleted, err
		}
		deleted++
	}
	return deleted, nil
}
