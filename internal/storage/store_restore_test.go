package storage

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"pzm/internal/domain"
)

func TestFsBackupStore_Restore_FolderBackup_OverwritesTarget(t *testing.T) {
	saveDir := t.TempDir()
	os.WriteFile(filepath.Join(saveDir, "map_meta.bin"), []byte("v1"), 0o644)

	backupRoot := t.TempDir()
	s := NewFsBackupStore(backupRoot)
	save := domain.Save{ID: "restore-test", Name: "S", GameMode: "A", Path: saveDir}

	b, err := s.Create(save, CreateOpts{Format: domain.BackupFormatFolder, Kind: domain.BackupKindManual})
	if err != nil {
		t.Fatal(err)
	}

	// user "gra" — modyfikuje save
	os.WriteFile(filepath.Join(saveDir, "map_meta.bin"), []byte("v2-changed"), 0o644)
	os.WriteFile(filepath.Join(saveDir, "newfile.txt"), []byte("trash"), 0o644)

	// restore — powinno wrócić do v1
	if err := s.Restore(b, saveDir); err != nil {
		t.Fatalf("restore: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(saveDir, "map_meta.bin"))
	if err != nil || string(data) != "v1" {
		t.Errorf("expected map_meta.bin='v1', got %q (err=%v)", data, err)
	}
	if _, err := os.Stat(filepath.Join(saveDir, "newfile.txt")); !os.IsNotExist(err) {
		t.Errorf("expected newfile.txt to be removed by restore")
	}
}

func TestFsBackupStore_Restore_ZipBackup_OverwritesTarget(t *testing.T) {
	saveDir := t.TempDir()
	os.WriteFile(filepath.Join(saveDir, "map_meta.bin"), []byte("v1"), 0o644)

	backupRoot := t.TempDir()
	s := NewFsBackupStore(backupRoot)
	save := domain.Save{ID: "zip-restore", Name: "S", GameMode: "A", Path: saveDir}

	b, err := s.Create(save, CreateOpts{Format: domain.BackupFormatZip, Kind: domain.BackupKindManual})
	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(filepath.Join(saveDir, "map_meta.bin"), []byte("CHANGED"), 0o644)

	if err := s.Restore(b, saveDir); err != nil {
		t.Fatalf("restore: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(saveDir, "map_meta.bin"))
	if string(data) != "v1" {
		t.Errorf("expected v1, got %q", data)
	}
}

func TestFsBackupStore_Restore_ZipSlipProtection(t *testing.T) {
	// build a malicious ZIP that contains a path with ../../
	zipPath := filepath.Join(t.TempDir(), "malicious.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := newTestZipWriter(t, f)
	addZipEntry(t, zw, "../../escaped.txt", []byte("pwned"))
	zw.Close()
	f.Close()

	target := filepath.Join(t.TempDir(), "target")

	err = unzipTo(zipPath, target)
	if err == nil {
		t.Error("expected error from zip-slip protection")
	}
}

func newTestZipWriter(t *testing.T, w *os.File) *zip.Writer {
	return zip.NewWriter(w)
}

func addZipEntry(t *testing.T, zw *zip.Writer, name string, data []byte) {
	t.Helper()
	h := &zip.FileHeader{Name: name, Method: zip.Deflate}
	fw, err := zw.CreateHeader(h)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write(data); err != nil {
		t.Fatal(err)
	}
}
