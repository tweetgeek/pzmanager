package ui

import (
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"pzm/internal/config"
)

func TestNewSettingsView_DoesNotCrash(t *testing.T) {
	fyneApp := test.NewApp()
	defer fyneApp.Quit()
	win := fyneApp.NewWindow("t")
	win.Show()
	defer win.Close()

	cfg := &config.Settings{
		SavesPath:        filepath.FromSlash("/tmp"),
		AutoBackupRetain: 5,
		DefaultFormat:    "zip",
	}
	svc := &Services{Cfg: cfg, CfgPath: filepath.Join(t.TempDir(), "settings.json")}

	view := NewSettingsView(svc, func(fyne.CanvasObject) {})
	if view == nil {
		t.Fatal("view nil")
	}
}
