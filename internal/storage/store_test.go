package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"pzm/internal/domain"
)

func TestFsBackupStore_List_EmptyDir_ReturnsEmpty(t *testing.T) {
	root := t.TempDir()
	s := NewFsBackupStore(root)
	backups, err := s.List("save-id-123")
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 0 {
		t.Errorf("expected 0 backups, got %d", len(backups))
	}
}

func TestFsBackupStore_Create_FolderFormat_CopiesContents(t *testing.T) {
	saveDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(saveDir, "map_meta.bin"), []byte("save data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(saveDir, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(saveDir, "subdir", "players.db"), []byte("players"), 0o644); err != nil {
		t.Fatal(err)
	}

	backupRoot := t.TempDir()
	s := NewFsBackupStore(backupRoot)

	save := domain.Save{
		ID:       "abc123",
		Name:     "Survivor_01",
		GameMode: "Apocalypse",
		Path:     saveDir,
	}
	backup, err := s.Create(save, CreateOpts{
		Format: domain.BackupFormatFolder,
		Kind:   domain.BackupKindManual,
		Label:  "before raid",
	})
	if err != nil {
		t.Fatal(err)
	}
	if backup.ID == "" {
		t.Error("expected non-empty backup ID")
	}
	if backup.SaveID != "abc123" {
		t.Errorf("SaveID = %q, want abc123", backup.SaveID)
	}
	if backup.Format != domain.BackupFormatFolder {
		t.Error("expected Format=Folder")
	}
	if backup.Kind != domain.BackupKindManual {
		t.Error("expected Kind=Manual")
	}
	if time.Since(backup.CreatedAt) > time.Minute {
		t.Error("CreatedAt seems wrong")
	}

	copiedRoot := backup.Path
	copiedMeta := filepath.Join(copiedRoot, "map_meta.bin")
	if data, err := os.ReadFile(copiedMeta); err != nil || string(data) != "save data" {
		t.Errorf("map_meta.bin not copied correctly: data=%q err=%v", data, err)
	}
	copiedPlayers := filepath.Join(copiedRoot, "subdir", "players.db")
	if data, err := os.ReadFile(copiedPlayers); err != nil || string(data) != "players" {
		t.Errorf("players.db not copied correctly: data=%q err=%v", data, err)
	}

	metaPath := filepath.Join(backupRoot, save.ID, backup.ID+".json")
	if _, err := os.Stat(metaPath); err != nil {
		t.Errorf("expected metadata file at %s, got err=%v", metaPath, err)
	}
}

func TestFsBackupStore_Create_ThenList_ReturnsBackup(t *testing.T) {
	saveDir := t.TempDir()
	os.WriteFile(filepath.Join(saveDir, "map_meta.bin"), []byte("x"), 0o644)

	backupRoot := t.TempDir()
	s := NewFsBackupStore(backupRoot)
	save := domain.Save{ID: "save-1", Name: "S1", GameMode: "Apoc", Path: saveDir}

	_, err := s.Create(save, CreateOpts{Format: domain.BackupFormatFolder, Kind: domain.BackupKindAuto})
	if err != nil {
		t.Fatal(err)
	}

	backups, err := s.List("save-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 1 {
		t.Fatalf("expected 1 backup, got %d", len(backups))
	}
}

func TestFsBackupStore_Delete_RemovesArchiveAndMetadata(t *testing.T) {
	saveDir := t.TempDir()
	os.WriteFile(filepath.Join(saveDir, "map_meta.bin"), []byte("x"), 0o644)

	backupRoot := t.TempDir()
	s := NewFsBackupStore(backupRoot)
	save := domain.Save{ID: "save-del", Name: "S", GameMode: "A", Path: saveDir}

	b, err := s.Create(save, CreateOpts{Format: domain.BackupFormatFolder, Kind: domain.BackupKindManual})
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Delete(b.ID); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(b.Path); !os.IsNotExist(err) {
		t.Errorf("archive should be removed; stat err=%v", err)
	}
	metaPath := filepath.Join(backupRoot, save.ID, b.ID+".json")
	if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
		t.Errorf("metadata should be removed; stat err=%v", err)
	}
}

func TestFsBackupStore_Delete_NotFound_ReturnsError(t *testing.T) {
	s := NewFsBackupStore(t.TempDir())
	if err := s.Delete("does-not-exist-id"); err == nil {
		t.Error("expected error for non-existent backup")
	}
}

func TestFsBackupStore_List_SkipsOrphanedMetadata(t *testing.T) {
	backupRoot := t.TempDir()
	saveDir := filepath.Join(backupRoot, "savex")
	os.MkdirAll(saveDir, 0o755)

	// Valid JSON metadata pointing at a non-existent archive. We marshal via
	// encoding/json so the path is properly escaped on Windows (raw backslashes
	// in a JSON string would produce invalid escape sequences and route through
	// the "corrupted metadata" branch instead of the "missing archive" branch
	// we intend to test.
	metaPath := filepath.Join(saveDir, "20260506-120000-deadbeef.json")
	meta := domain.Backup{
		ID:        "20260506-120000-deadbeef",
		SaveID:    "savex",
		SaveName:  "X",
		CreatedAt: time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC),
		Kind:      domain.BackupKindManual,
		Format:    domain.BackupFormatFolder,
		Path:      filepath.Join(saveDir, "nonexistent-archive"),
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(metaPath, metaJSON, 0o644)

	s := NewFsBackupStore(backupRoot)
	backups, err := s.List("savex")
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 0 {
		t.Errorf("expected 0 backups (orphan filtered), got %d", len(backups))
	}
}
