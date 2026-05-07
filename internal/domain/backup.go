package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type BackupKind int

const (
	BackupKindAuto BackupKind = iota + 1
	BackupKindManual
)

func (k BackupKind) String() string {
	switch k {
	case BackupKindAuto:
		return "auto"
	case BackupKindManual:
		return "manual"
	default:
		return "unknown"
	}
}

func ParseBackupKind(s string) (BackupKind, error) {
	switch strings.ToLower(s) {
	case "auto":
		return BackupKindAuto, nil
	case "manual":
		return BackupKindManual, nil
	default:
		return 0, fmt.Errorf("unknown backup kind: %q", s)
	}
}

func (k BackupKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

func (k *BackupKind) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseBackupKind(s)
	if err != nil {
		return err
	}
	*k = parsed
	return nil
}

type BackupFormat int

const (
	BackupFormatZip BackupFormat = iota + 1
	BackupFormatFolder
)

func (f BackupFormat) String() string {
	switch f {
	case BackupFormatZip:
		return "zip"
	case BackupFormatFolder:
		return "folder"
	default:
		return "unknown"
	}
}

func ParseBackupFormat(s string) (BackupFormat, error) {
	switch strings.ToLower(s) {
	case "zip":
		return BackupFormatZip, nil
	case "folder":
		return BackupFormatFolder, nil
	default:
		return 0, fmt.Errorf("unknown backup format: %q", s)
	}
}

func (f BackupFormat) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

func (f *BackupFormat) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseBackupFormat(s)
	if err != nil {
		return err
	}
	*f = parsed
	return nil
}

type Backup struct {
	ID          string       `json:"id"`
	SaveID      string       `json:"save_id"`
	SaveName    string       `json:"save_name"`
	CreatedAt   time.Time    `json:"created_at"`
	Kind        BackupKind   `json:"kind"`
	Format      BackupFormat `json:"format"`
	Path        string       `json:"path"`
	SizeBytes   int64        `json:"size_bytes"`
	Label       string       `json:"label,omitempty"`
	GameVersion string       `json:"game_version,omitempty"`
	SaveHash    string       `json:"save_hash,omitempty"`
}
