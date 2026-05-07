package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_NonExistent_ReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	s, err := Load(filepath.Join(dir, "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if s.Version != 1 {
		t.Errorf("expected Version=1, got %d", s.Version)
	}
	if s.AutoBackupRetain != 5 {
		t.Errorf("expected AutoBackupRetain=5, got %d", s.AutoBackupRetain)
	}
	if !s.AutoBackup {
		t.Errorf("expected AutoBackup=true by default")
	}
	if s.DefaultFormat != "zip" {
		t.Errorf("expected DefaultFormat=zip, got %q", s.DefaultFormat)
	}
}

func TestSave_ThenLoad_RoundTrips(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "settings.json")

	orig := &Settings{
		Version:          1,
		SavesPath:        "/foo/saves",
		BackupRoot:       "/foo/backups",
		AutoBackup:       false,
		AutoBackupRetain: 12,
		DefaultFormat:    "folder",
	}
	if err := orig.Save(p); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.SavesPath != "/foo/saves" {
		t.Errorf("SavesPath = %q", loaded.SavesPath)
	}
	if loaded.AutoBackupRetain != 12 {
		t.Errorf("Retain = %d", loaded.AutoBackupRetain)
	}
	if loaded.AutoBackup != false {
		t.Errorf("AutoBackup = %v", loaded.AutoBackup)
	}
	if loaded.DefaultFormat != "folder" {
		t.Errorf("DefaultFormat = %q", loaded.DefaultFormat)
	}
}

func TestSave_Atomic_NoCorruptionFromPartialWrite(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "settings.json")
	s := &Settings{Version: 1, AutoBackupRetain: 1, DefaultFormat: "zip"}
	if err := s.Save(p); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 || entries[0].Name() != "settings.json" {
		t.Errorf("expected single settings.json, got: %v", entries)
	}
}
