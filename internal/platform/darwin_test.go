//go:build darwin

package platform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDarwin_DefaultPaths(t *testing.T) {
	home, _ := os.UserHomeDir()
	d := New()
	if want := filepath.Join(home, "Zomboid", "Saves"); d.DefaultSavesPath() != want {
		t.Errorf("Saves: got %q, want %q", d.DefaultSavesPath(), want)
	}
	if !strings.Contains(d.DefaultGameInstallPath(), "ProjectZomboid") {
		t.Errorf("InstallPath suspicious: %q", d.DefaultGameInstallPath())
	}
	if !strings.Contains(d.DefaultConfigPath(), "PZManager") {
		t.Errorf("ConfigPath suspicious: %q", d.DefaultConfigPath())
	}
	if d.DefaultBackupPath() == "" {
		t.Error("BackupPath empty")
	}
}
