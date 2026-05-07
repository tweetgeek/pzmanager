//go:build linux

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"pzm/internal/domain"
)

type linuxPlatform struct{}

// New zwraca implementację Platform dla Linuksa.
func New() Platform { return &linuxPlatform{} }

func (l *linuxPlatform) Identity() domain.Platform { return domain.PlatformLinux }

func (l *linuxPlatform) DefaultSavesPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Zomboid", "Saves")
}

func (l *linuxPlatform) DefaultGameInstallPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".steam", "steam", "steamapps", "common", "ProjectZomboid")
}

func (l *linuxPlatform) DefaultBackupPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "PZManager", "Backups")
}

func (l *linuxPlatform) DefaultConfigPath() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "pz-manager")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "pz-manager")
}

func (l *linuxPlatform) DetectGame() (domain.GameInstance, error) {
	install := l.DefaultGameInstallPath()
	if _, err := os.Stat(install); err != nil {
		return domain.GameInstance{}, fmt.Errorf("game not found at %s: %w", install, err)
	}
	return domain.GameInstance{
		Platform:    domain.PlatformLinux,
		InstallPath: install,
		SavesPath:   l.DefaultSavesPath(),
		Source:      domain.SourceSteam,
		SteamAppID:  "108600",
	}, nil
}

func (l *linuxPlatform) LaunchGame(inst domain.GameInstance, steamCommand []string) error {
	if len(steamCommand) > 0 {
		cmd := exec.Command(steamCommand[0], steamCommand[1:]...)
		return cmd.Start()
	}
	if inst.Source == domain.SourceSteam && inst.SteamAppID != "" {
		return exec.Command("xdg-open", "steam://run/"+inst.SteamAppID).Start()
	}
	exe := filepath.Join(inst.InstallPath, "projectzomboid")
	return exec.Command(exe).Start()
}

func (l *linuxPlatform) InstallShortcut(target string) error {
	home, _ := os.UserHomeDir()
	desktop := filepath.Join(home, "Desktop")
	if _, err := os.Stat(desktop); err != nil {
		desktop = filepath.Join(home, ".local", "share", "applications")
		if err := os.MkdirAll(desktop, 0o755); err != nil {
			return err
		}
	}
	content := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=Project Zomboid (via Manager)
Exec="%s" --auto-launch
Terminal=false
Categories=Game;
`, target)

	return os.WriteFile(filepath.Join(desktop, "pz-manager.desktop"), []byte(content), 0o755)
}
