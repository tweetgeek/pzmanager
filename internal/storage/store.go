package storage

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"pzm/internal/domain"
)

// BackupStore zarządza backupami save'ów na dysku.
//
// Implementacja nie używa wewnętrznego locking'u — zakładamy, że wywołania
// dla tego samego save'a są serializowane przez warstwę wyżej (UI / launcher).
// W aplikacji single-user desktopowej spełnione naturalnie.
//
// PruneAuto może zwrócić błąd po częściowym usunięciu — `deleted` zawiera
// liczbę faktycznie skasowanych elementów do tego momentu. Caller decyduje czy
// retry, czy zaakceptować częściowy stan.
type BackupStore interface {
	List(saveID string) ([]domain.Backup, error)
	Create(save domain.Save, opts CreateOpts) (domain.Backup, error)
	Restore(backup domain.Backup, target string) error
	Delete(backupID string) error
	PruneAuto(saveID string, keep int) (deleted int, err error)
}

type CreateOpts struct {
	Format domain.BackupFormat
	Kind   domain.BackupKind
	Label  string
}

type fsBackupStore struct {
	root string
}

func NewFsBackupStore(root string) BackupStore {
	return &fsBackupStore{root: root}
}

func (s *fsBackupStore) List(saveID string) ([]domain.Backup, error) {
	saveDir := filepath.Join(s.root, saveID)
	entries, err := os.ReadDir(saveDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read backup dir: %w", err)
	}

	var backups []domain.Backup
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		b, err := s.readMeta(saveDir, e.Name())
		if err != nil {
			log.Printf("backup list: skipping corrupted metadata %s: %v", e.Name(), err)
			continue
		}
		if _, err := os.Stat(b.Path); err != nil {
			log.Printf("backup list: skipping orphaned metadata %s (archive missing: %v)", e.Name(), err)
			continue
		}
		backups = append(backups, b)
	}
	sort.Slice(backups, func(i, j int) bool { return backups[i].CreatedAt.After(backups[j].CreatedAt) })
	return backups, nil
}

func (s *fsBackupStore) readMeta(saveDir, fileName string) (domain.Backup, error) {
	raw, err := os.ReadFile(filepath.Join(saveDir, fileName))
	if err != nil {
		return domain.Backup{}, err
	}
	var b domain.Backup
	if err := json.Unmarshal(raw, &b); err != nil {
		return domain.Backup{}, err
	}
	return b, nil
}

func (s *fsBackupStore) writeMeta(saveDir string, b domain.Backup) error {
	if err := os.MkdirAll(saveDir, 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	metaPath := filepath.Join(saveDir, b.ID+".json")
	return atomicWrite(metaPath, raw)
}

func (s *fsBackupStore) Delete(backupID string) error {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		saveDir := filepath.Join(s.root, e.Name())
		metaPath := filepath.Join(saveDir, backupID+".json")
		if _, err := os.Stat(metaPath); err != nil {
			continue
		}
		b, err := s.readMeta(saveDir, backupID+".json")
		if err != nil {
			return err
		}
		if err := s.assertWithinRoot(b.Path); err != nil {
			return fmt.Errorf("backup metadata invalid: %w", err)
		}
		if err := os.RemoveAll(b.Path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return os.Remove(metaPath)
	}
	return fmt.Errorf("backup %q not found", backupID)
}

// assertWithinRoot guards against malformed metadata pointing outside the
// backup root (path traversal). Used before any destructive op based on
// b.Path read from a JSON file.
func (s *fsBackupStore) assertWithinRoot(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	absRoot, err := filepath.Abs(s.root)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path %q escapes backup root", path)
	}
	return nil
}


func newBackupID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic(fmt.Sprintf("crypto/rand.Read failed: %v", err))
	}
	return time.Now().UTC().Format("20060102-150405") + "-" + hex.EncodeToString(buf[:])
}

func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}
