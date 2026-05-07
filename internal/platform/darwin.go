//go:build darwin

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"pzm/internal/domain"
)

type darwinPlatform struct{}

// New zwraca implementację Platform dla macOS.
func New() Platform { return &darwinPlatform{} }

func (d *darwinPlatform) Identity() domain.Platform { return domain.PlatformMacOS }

func (d *darwinPlatform) DefaultSavesPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Zomboid", "Saves")
}

func (d *darwinPlatform) DefaultGameInstallPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "Application Support", "Steam", "steamapps", "common", "ProjectZomboid")
}

func (d *darwinPlatform) DefaultBackupPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Documents", "PZManager", "Backups")
}

func (d *darwinPlatform) DefaultConfigPath() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "PZManager")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "Application Support", "PZManager")
}

func (d *darwinPlatform) DetectGame() (domain.GameInstance, error) {
	install := d.DefaultGameInstallPath()
	if _, err := os.Stat(install); err != nil {
		return domain.GameInstance{}, fmt.Errorf("game not found at %s: %w", install, err)
	}
	return domain.GameInstance{
		Platform:    domain.PlatformMacOS,
		InstallPath: install,
		SavesPath:   d.DefaultSavesPath(),
		Source:      domain.SourceSteam,
		SteamAppID:  "108600",
	}, nil
}

func (d *darwinPlatform) LaunchGame(inst domain.GameInstance, steamCommand []string) error {
	if len(steamCommand) > 0 {
		cmd := exec.Command(steamCommand[0], steamCommand[1:]...)
		return cmd.Start()
	}
	if inst.Source == domain.SourceSteam && inst.SteamAppID != "" {
		return exec.Command("open", "steam://run/"+inst.SteamAppID).Start()
	}
	return exec.Command("open", "-a", "Project Zomboid").Start()
}

func (d *darwinPlatform) InstallShortcut(target string) error {
	home, _ := os.UserHomeDir()
	desktop := filepath.Join(home, "Desktop")
	scriptPath := filepath.Join(desktop, "Project Zomboid Manager.command")
	content := fmt.Sprintf("#!/bin/bash\n\"%s\" --auto-launch\n", target)
	return os.WriteFile(scriptPath, []byte(content), 0o755)
}
