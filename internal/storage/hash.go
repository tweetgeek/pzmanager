package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// HashSavePath returns a deterministic short hash of the save's absolute path.
// Used as a stable ID in domain.Save.
func HashSavePath(absPath string) string {
	sum := sha256.Sum256([]byte(filepath.Clean(absPath)))
	return hex.EncodeToString(sum[:])[:16]
}

// HashSaveContents computes a hash of the save's contents based on sizes
// and mtime of all files. Fast, deterministic, detects real changes.
func HashSaveContents(saveDir string) (string, error) {
	info, err := os.Stat(saveDir)
	if err != nil {
		return "", fmt.Errorf("save dir: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("not a directory: %q", saveDir)
	}

	h := sha256.New()
	type entry struct {
		rel  string
		size int64
		mod  int64
	}
	var entries []entry

	err = filepath.WalkDir(saveDir, func(path string, d os.DirEntry, err error) error {
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
		rel, err := filepath.Rel(saveDir, path)
		if err != nil {
			return err
		}
		entries = append(entries, entry{rel: rel, size: info.Size(), mod: info.ModTime().UnixNano()})
		return nil
	})
	if err != nil {
		return "", err
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].rel < entries[j].rel })

	for _, e := range entries {
		fmt.Fprintf(h, "%s\x00%d\x00%d\n", e.rel, e.size, e.mod)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
