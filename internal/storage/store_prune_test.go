package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"pzm/internal/domain"
)

func TestFsBackupStore_PruneAuto_KeepsNewestN_AndAllManual(t *testing.T) {
	saveDir := t.TempDir()
	os.WriteFile(filepath.Join(saveDir, "map_meta.bin"), []byte("x"), 0o644)

	backupRoot := t.TempDir()
	s := NewFsBackupStore(backupRoot)
	save := domain.Save{ID: "prune-test", Name: "S", GameMode: "A", Path: saveDir}

	// 5 auto + 2 manual; vary timestamps
	for i := 0; i < 5; i++ {
		b, err := s.Create(save, CreateOpts{Format: domain.BackupFormatFolder, Kind: domain.BackupKindAuto})
		if err != nil {
			t.Fatal(err)
		}
		forceTime(t, backupRoot, save.ID, b.ID, time.Now().Add(-time.Duration(10-i)*time.Minute))
	}
	for i := 0; i < 2; i++ {
		b, err := s.Create(save, CreateOpts{Format: domain.BackupFormatFolder, Kind: domain.BackupKindManual})
		if err != nil {
			t.Fatal(err)
		}
		forceTime(t, backupRoot, save.ID, b.ID, time.Now().Add(-time.Duration(20-i)*time.Hour))
	}

	deleted, err := s.PruneAuto(save.ID, 2)
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 3 {
		t.Errorf("expected 3 deleted (5 auto - 2 keep), got %d", deleted)
	}

	backups, _ := s.List(save.ID)
	autoCount, manualCount := 0, 0
	for _, b := range backups {
		if b.Kind == domain.BackupKindAuto {
			autoCount++
		}
		if b.Kind == domain.BackupKindManual {
			manualCount++
		}
	}
	if autoCount != 2 {
		t.Errorf("expected 2 auto remaining, got %d", autoCount)
	}
	if manualCount != 2 {
		t.Errorf("expected 2 manual remaining (untouched), got %d", manualCount)
	}
}

func TestFsBackupStore_PruneAuto_NothingToDoWhenWithinLimit(t *testing.T) {
	saveDir := t.TempDir()
	os.WriteFile(filepath.Join(saveDir, "map_meta.bin"), []byte("x"), 0o644)
	s := NewFsBackupStore(t.TempDir())
	save := domain.Save{ID: "small", Name: "S", GameMode: "A", Path: saveDir}

	for i := 0; i < 2; i++ {
		if _, err := s.Create(save, CreateOpts{Format: domain.BackupFormatFolder, Kind: domain.BackupKindAuto}); err != nil {
			t.Fatal(err)
		}
	}

	deleted, err := s.PruneAuto(save.ID, 5)
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deleted (under limit), got %d", deleted)
	}
}

// forceTime overrides CreatedAt in metadata for retention tests.
func forceTime(t *testing.T, root, saveID, backupID string, ts time.Time) {
	t.Helper()
	saveDir := filepath.Join(root, saveID)
	fbs := &fsBackupStore{root: root}
	b, err := fbs.readMeta(saveDir, backupID+".json")
	if err != nil {
		t.Fatal(err)
	}
	b.CreatedAt = ts
	if err := fbs.writeMeta(saveDir, b); err != nil {
		t.Fatal(err)
	}
}
