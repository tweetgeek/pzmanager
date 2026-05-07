package storage

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"pzm/internal/domain"
)

func TestFsBackupStore_Create_ZipFormat_ProducesValidArchive(t *testing.T) {
	saveDir := t.TempDir()
	os.WriteFile(filepath.Join(saveDir, "map_meta.bin"), []byte("metadata"), 0o644)
	os.MkdirAll(filepath.Join(saveDir, "sub"), 0o755)
	os.WriteFile(filepath.Join(saveDir, "sub", "players.db"), []byte("playerdata"), 0o644)

	backupRoot := t.TempDir()
	s := NewFsBackupStore(backupRoot)
	save := domain.Save{ID: "zip-test", Name: "S", GameMode: "Apoc", Path: saveDir}

	b, err := s.Create(save, CreateOpts{
		Format: domain.BackupFormatZip,
		Kind:   domain.BackupKindManual,
	})
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(b.Path) != ".zip" {
		t.Errorf("expected .zip extension, got %s", b.Path)
	}
	if b.SizeBytes <= 0 {
		t.Errorf("expected non-zero size, got %d", b.SizeBytes)
	}

	r, err := zip.OpenReader(b.Path)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer r.Close()

	found := map[string][]byte{}
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, _ := io.ReadAll(rc)
		rc.Close()
		found[f.Name] = data
	}

	if string(found["map_meta.bin"]) != "metadata" {
		t.Errorf("map_meta.bin content mismatch: %q", found["map_meta.bin"])
	}
	if string(found["sub/players.db"]) != "playerdata" {
		t.Errorf("sub/players.db content mismatch: %q", found["sub/players.db"])
	}
}
