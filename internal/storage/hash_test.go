package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHashSavePath_DeterministicForSamePath(t *testing.T) {
	h1 := HashSavePath("/foo/bar/Survivor_01")
	h2 := HashSavePath("/foo/bar/Survivor_01")
	if h1 != h2 {
		t.Errorf("hashes differ: %s vs %s", h1, h2)
	}
	if len(h1) < 8 {
		t.Errorf("hash too short: %s", h1)
	}
}

func TestHashSavePath_DifferentForDifferentPaths(t *testing.T) {
	if HashSavePath("/a") == HashSavePath("/b") {
		t.Error("hashes for different paths should differ")
	}
}

func TestHashSaveContents_SameFiles_SameHash(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "map_meta.bin"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	h1, err := HashSaveContents(dir)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := HashSaveContents(dir)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Errorf("repeated hash differs: %s vs %s", h1, h2)
	}
}

func TestHashSaveContents_ChangedFile_ChangedHash(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "map_meta.bin")
	os.WriteFile(p, []byte("v1"), 0o644)
	h1, _ := HashSaveContents(dir)
	os.WriteFile(p, []byte("v2"), 0o644)
	h2, _ := HashSaveContents(dir)
	if h1 == h2 {
		t.Error("expected hash to change after file content change")
	}
}

func TestHashSaveContents_NonExistentDir_ReturnsError(t *testing.T) {
	_, err := HashSaveContents("/does/not/exist/zzz_pzsm_test")
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}

func TestHashSaveContents_FileNotDir_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "afile")
	os.WriteFile(p, []byte("x"), 0o644)
	_, err := HashSaveContents(p) // pass file path, not dir
	if err == nil {
		t.Error("expected error when path is a file not directory")
	}
}
