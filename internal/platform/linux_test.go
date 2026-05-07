//go:build linux

package platform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinux_DefaultPaths(t *testing.T) {
	home, _ := os.UserHomeDir()
	l := New()
	if want := filepath.Join(home, "Zomboid", "Saves"); l.DefaultSavesPath() != want {
		t.Errorf("Saves: got %q, want %q", l.DefaultSavesPath(), want)
	}
	if !strings.Contains(l.DefaultGameInstallPath(), "ProjectZomboid") {
		t.Errorf("InstallPath suspicious: %q", l.DefaultGameInstallPath())
	}
	if !strings.Contains(l.DefaultConfigPath(), "pz-manager") {
		t.Errorf("ConfigPath suspicious: %q", l.DefaultConfigPath())
	}
	if l.DefaultBackupPath() == "" {
		t.Error("BackupPath empty")
	}
}
