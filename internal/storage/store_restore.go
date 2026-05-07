package storage

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"pzm/internal/domain"
)

// Restore odtwarza save z backupu w sposób atomowy:
// 1) wypakuj/skopiuj backup do tempu obok target
// 2) przenieś bieżący target na bok (jeśli istnieje)
// 3) os.Rename(temp, target)
// 4) skasuj odsunięty stary target
//
// Daje gwarancję: przerwanie w trakcie nie zostawia uszkodzonego save'a.
func (s *fsBackupStore) Restore(b domain.Backup, target string) error {
	if err := s.assertWithinRoot(b.Path); err != nil {
		return fmt.Errorf("backup path: %w", err)
	}

	parent := filepath.Dir(target)
	base := filepath.Base(target)
	tempTarget := filepath.Join(parent, base+".restore-tmp")

	// sprzątanie po poprzedniej nieudanej próbie
	_ = os.RemoveAll(tempTarget)

	switch b.Format {
	case domain.BackupFormatFolder:
		if _, err := copyDir(b.Path, tempTarget); err != nil {
			os.RemoveAll(tempTarget)
			return fmt.Errorf("copy folder backup: %w", err)
		}
	case domain.BackupFormatZip:
		if err := unzipTo(b.Path, tempTarget); err != nil {
			os.RemoveAll(tempTarget)
			return fmt.Errorf("unzip: %w", err)
		}
	default:
		return fmt.Errorf("unknown backup format: %v", b.Format)
	}

	// atomic swap
	backupOld := filepath.Join(parent, base+".restore-old")
	_ = os.RemoveAll(backupOld)
	if _, err := os.Stat(target); err == nil {
		if err := os.Rename(target, backupOld); err != nil {
			os.RemoveAll(tempTarget)
			return fmt.Errorf("move existing target aside: %w", err)
		}
	}
	if err := os.Rename(tempTarget, target); err != nil {
		// Recovery: only safe if backupOld exists (which it does iff we moved
		// something aside). Try to put the old save back. If recovery itself
		// fails, return both errors so caller knows the state is degraded.
		if _, statErr := os.Stat(backupOld); statErr == nil {
			if rollbackErr := os.Rename(backupOld, target); rollbackErr != nil {
				return fmt.Errorf("swap failed: %w; rollback also failed: %v (manual intervention needed)", err, rollbackErr)
			}
		}
		return fmt.Errorf("swap restored save into place: %w", err)
	}
	_ = os.RemoveAll(backupOld)
	return nil
}

func unzipTo(zipPath, dst string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}

	cleanDst := filepath.Clean(dst)
	for _, f := range r.File {
		// Convert ZIP-internal forward slashes to OS-native, then resolve.
		target := filepath.Join(dst, filepath.FromSlash(f.Name))

		// Zip-slip protection: verify target is within dst.
		cleanTarget := filepath.Clean(target)
		if cleanTarget != cleanDst && !strings.HasPrefix(cleanTarget, cleanDst+string(os.PathSeparator)) {
			return fmt.Errorf("zip slip detected: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			out.Close()
			return err
		}
		_, copyErr := io.Copy(out, rc)
		rc.Close()
		out.Close()
		if copyErr != nil {
			os.Remove(target) // remove partially-written file
			return copyErr
		}
	}
	return nil
}
