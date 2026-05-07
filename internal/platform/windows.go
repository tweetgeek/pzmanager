//go:build windows

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"pzm/internal/domain"
)

type windowsPlatform struct{}

// New zwraca implementację Platform dla Windows.
func New() Platform { return &windowsPlatform{} }

func (w *windowsPlatform) Identity() domain.Platform { return domain.PlatformWindows }

func (w *windowsPlatform) DefaultSavesPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Zomboid", "Saves")
}

func (w *windowsPlatform) DefaultGameInstallPath() string {
	candidates := []string{
		`C:\Program Files (x86)\Steam\steamapps\common\ProjectZomboid`,
		`C:\Program Files\Steam\steamapps\common\ProjectZomboid`,
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return candidates[0]
}

func (w *windowsPlatform) DefaultBackupPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Documents", "PZManager", "Backups")
}

func (w *windowsPlatform) DefaultConfigPath() string {
	if dir := os.Getenv("APPDATA"); dir != "" {
		return filepath.Join(dir, "PZManager")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "AppData", "Roaming", "PZManager")
}

func (w *windowsPlatform) DetectGame() (domain.GameInstance, error) {
	install := w.DefaultGameInstallPath()
	if _, err := os.Stat(install); err != nil {
		return domain.GameInstance{}, fmt.Errorf("game not found: %w", err)
	}
	return domain.GameInstance{
		Platform:    domain.PlatformWindows,
		InstallPath: install,
		SavesPath:   w.DefaultSavesPath(),
		Source:      domain.SourceSteam,
		SteamAppID:  "108600",
	}, nil
}

func (w *windowsPlatform) LaunchGame(inst domain.GameInstance, steamCommand []string) error {
	if len(steamCommand) > 0 {
		cmd := exec.Command(steamCommand[0], steamCommand[1:]...)
		return cmd.Start()
	}
	if inst.Source == domain.SourceSteam && inst.SteamAppID != "" {
		return exec.Command("cmd", "/c", "start", "steam://run/"+inst.SteamAppID).Start()
	}
	exe := filepath.Join(inst.InstallPath, "ProjectZomboid64.exe")
	return exec.Command(exe).Start()
}

// InstallShortcut tworzy .lnk na pulpicie wskazujący na target z argumentem
// --auto-launch. Używa PowerShell + WSH, nie wymaga zewnętrznych zależności.
func (w *windowsPlatform) InstallShortcut(target string) error {
	home, _ := os.UserHomeDir()
	desktop := filepath.Join(home, "Desktop")
	lnkPath := filepath.Join(desktop, "Project Zomboid Manager.lnk")

	psScript := fmt.Sprintf(`
$WshShell = New-Object -ComObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut("%s")
$Shortcut.TargetPath = "%s"
$Shortcut.Arguments = "--auto-launch"
$Shortcut.WorkingDirectory = "%s"
$Shortcut.Save()
`, lnkPath, target, filepath.Dir(target))

	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("powershell create shortcut: %w (output: %s)", err, output)
	}
	return nil
}
