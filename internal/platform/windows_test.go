//go:build windows

package platform

import "testing"

func TestWindows_DefaultPaths(t *testing.T) {
	w := New()
	if w.DefaultSavesPath() == "" {
		t.Error("Saves empty")
	}
	if w.DefaultGameInstallPath() == "" {
		t.Error("Install empty")
	}
	if w.DefaultBackupPath() == "" {
		t.Error("Backup empty")
	}
	if w.DefaultConfigPath() == "" {
		t.Error("Config empty")
	}
}
