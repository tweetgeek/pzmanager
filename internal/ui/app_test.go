package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"

	"pzm/internal/config"
)

func TestApp_BuildsAndShowsSavesView(t *testing.T) {
	fyneApp := test.NewApp()
	defer fyneApp.Quit()

	services := &Services{Cfg: &config.Settings{}}
	a := NewApp(fyneApp, services)
	a.Show()

	if a.window.Content() == nil {
		t.Fatal("window content is nil")
	}
}

func TestNavSidebar_OnSelect_FiresCallback(t *testing.T) {
	var called string
	nav := NewNavSidebar(func(section string) {
		called = section
	})
	nav.Select("settings")
	if called != "settings" {
		t.Errorf("callback got %q, want settings", called)
	}
}
