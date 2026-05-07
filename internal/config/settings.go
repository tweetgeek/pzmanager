package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type LauncherInstalled struct {
	Shortcut    bool `json:"shortcut"`
	SteamCopied bool `json:"steam_copied"`
}

type Settings struct {
	Version           int               `json:"version"`
	SavesPath         string            `json:"saves_path,omitempty"`
	GameInstallPath   string            `json:"game_install_path,omitempty"`
	BackupRoot        string            `json:"backup_root,omitempty"`
	AutoBackup        bool              `json:"auto_backup"`
	AutoBackupRetain  int               `json:"auto_backup_retain"`
	DefaultFormat     string            `json:"default_format"`
	CloseOnLaunch     bool              `json:"close_on_launch"`
	LauncherInstalled LauncherInstalled `json:"launcher_installed"`
}

// Defaults zwraca settings z domyślnymi wartościami.
func Defaults() *Settings {
	return &Settings{
		Version:          1,
		AutoBackup:       true,
		AutoBackupRetain: 5,
		DefaultFormat:    "zip",
		CloseOnLaunch:    false,
	}
}

// Load wczytuje settings z pliku. Jeśli plik nie istnieje — zwraca Defaults().
func Load(path string) (*Settings, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Defaults(), nil
		}
		return nil, fmt.Errorf("read settings: %w", err)
	}
	var s Settings
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("parse settings: %w", err)
	}
	if s.Version == 0 {
		s.Version = 1
	}
	return &s, nil
}

// Save zapisuje settings atomowo (tmp + rename).
func (s *Settings) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}
