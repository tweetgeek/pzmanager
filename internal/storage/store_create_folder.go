package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"pzm/internal/domain"
)

func (s *fsBackupStore) Create(save domain.Save, opts CreateOpts) (domain.Backup, error) {
	backupID := newBackupID()
	saveBackupDir := filepath.Join(s.root, save.ID)
	if err := os.MkdirAll(saveBackupDir, 0o755); err != nil {
		return domain.Backup{}, fmt.Errorf("mkdir backup save dir: %w", err)
	}

	var (
		archivePath string
		sizeBytes   int64
		err         error
	)
	switch opts.Format {
	case domain.BackupFormatFolder:
		archivePath = filepath.Join(saveBackupDir, backupID)
		sizeBytes, err = copyDir(save.Path, archivePath)
	case domain.BackupFormatZip:
		archivePath = filepath.Join(saveBackupDir, backupID+".zip")
		sizeBytes, err = zipDir(save.Path, archivePath)
	default:
		return domain.Backup{}, fmt.Errorf("unsupported format: %v", opts.Format)
	}
	if err != nil {
		return domain.Backup{}, err
	}

	saveHash, _ := HashSaveContents(save.Path)

	b := domain.Backup{
		ID:          backupID,
		SaveID:      save.ID,
		SaveName:    save.Name,
		CreatedAt:   time.Now().UTC(),
		Kind:        opts.Kind,
		Format:      opts.Format,
		Path:        archivePath,
		SizeBytes:   sizeBytes,
		Label:       opts.Label,
		GameVersion: save.GameVersion,
		SaveHash:    saveHash,
	}
	if err := s.writeMeta(saveBackupDir, b); err != nil {
		return domain.Backup{}, err
	}
	return b, nil
}

func copyDir(src, dst string) (int64, error) {
	var total int64
	err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		n, err := copyFile(path, targetPath)
		total += n
		return err
	})
	return total, err
}

func copyFile(src, dst string) (int64, error) {
	in, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return 0, err
	}
	out, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	return io.Copy(out, in)
}
