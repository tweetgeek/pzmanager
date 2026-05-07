package main

import (
	"os"
	"path/filepath"
	"testing"

	"pzm/internal/config"
	"pzm/internal/domain"
	"pzm/internal/storage"
)

// TestSmoke_FullBackupRestoreCycle exercises the complete flow:
// scan saves → create backup → modify save → restore → verify original.
// Plus a config roundtrip.
func TestSmoke_FullBackupRestoreCycle(t *testing.T) {
	savesRoot := t.TempDir()
	saveDir := filepath.Join(savesRoot, "Apocalypse", "Survivor_01")
	if err := os.MkdirAll(saveDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(saveDir, "map_meta.bin"), []byte("game data"), 0o644); err != nil {
		t.Fatal(err)
	}

	backupRoot := t.TempDir()

	sc := storage.NewFsSaveScanner()
	saves, err := sc.Scan(savesRoot)
	if err != nil || len(saves) != 1 {
		t.Fatalf("scan: err=%v len=%d", err, len(saves))
	}

	st := storage.NewFsBackupStore(backupRoot)
	b, err := st.Create(saves[0], storage.CreateOpts{
		Format: domain.BackupFormatZip,
		Kind:   domain.BackupKindManual,
	})
	if err != nil {
		t.Fatal(err)
	}
	if b.SizeBytes <= 0 {
		t.Error("backup has zero size")
	}

	// user "plays" — modifies save
	if err := os.WriteFile(filepath.Join(saveDir, "map_meta.bin"), []byte("CHANGED"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := st.Restore(b, saves[0].Path); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(saveDir, "map_meta.bin"))
	if string(data) != "game data" {
		t.Errorf("expected restore to bring back 'game data', got %q", data)
	}

	// settings roundtrip
	cfgPath := filepath.Join(t.TempDir(), "settings.json")
	cfg := config.Defaults()
	cfg.SavesPath = savesRoot
	cfg.BackupRoot = backupRoot
	if err := cfg.Save(cfgPath); err != nil {
		t.Fatal(err)
	}
	loaded, err := config.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.SavesPath != savesRoot {
		t.Errorf("settings roundtrip lost SavesPath")
	}
}
