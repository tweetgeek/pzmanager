package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFsSaveScanner_Scan_FindsSavesInGameModeDirs(t *testing.T) {
	root := t.TempDir()

	mustMkdir(t, filepath.Join(root, "Apocalypse", "Survivor_01"))
	mustWrite(t, filepath.Join(root, "Apocalypse", "Survivor_01", "map_meta.bin"), []byte("data"))

	mustMkdir(t, filepath.Join(root, "Sandbox", "Custom_Build"))
	mustWrite(t, filepath.Join(root, "Sandbox", "Custom_Build", "map_meta.bin"), []byte("data2"))

	scanner := NewFsSaveScanner()
	saves, err := scanner.Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(saves) != 2 {
		t.Fatalf("expected 2 saves, got %d: %+v", len(saves), saves)
	}

	byName := map[string]bool{}
	for _, s := range saves {
		byName[s.Name] = true
		if s.ID == "" {
			t.Errorf("save %s has empty ID", s.Name)
		}
		if s.GameMode == "" {
			t.Errorf("save %s has empty GameMode", s.Name)
		}
	}
	if !byName["Survivor_01"] || !byName["Custom_Build"] {
		t.Errorf("missing expected saves, got: %v", byName)
	}
}

func TestFsSaveScanner_Scan_EmptyRoot_ReturnsEmptySlice(t *testing.T) {
	root := t.TempDir()
	scanner := NewFsSaveScanner()
	saves, err := scanner.Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(saves) != 0 {
		t.Errorf("expected 0 saves, got %d", len(saves))
	}
}

func TestFsSaveScanner_Scan_NonExistentRoot_ReturnsError(t *testing.T) {
	scanner := NewFsSaveScanner()
	_, err := scanner.Scan("/does/not/exist/zzz")
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}

// helpers
func mustMkdir(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}
func mustWrite(t *testing.T, p string, data []byte) {
	t.Helper()
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
