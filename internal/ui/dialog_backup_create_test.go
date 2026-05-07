package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"

	"pzm/internal/config"
	"pzm/internal/domain"
)

func TestShowBackupCreateDialog_DoesNotCrash(t *testing.T) {
	fyneApp := test.NewApp()
	defer fyneApp.Quit()
	win := fyneApp.NewWindow("test")
	win.Show()
	defer win.Close()

	svc := &Services{
		Cfg:   &config.Settings{DefaultFormat: "zip"},
		Store: &stubStore{},
	}

	ShowBackupCreateDialog(svc, domain.Save{ID: "x", Name: "X"}, func() {})
	// Pass if no panic.
}
