package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"

	"pzm/internal/config"
	"pzm/internal/domain"
)

func TestShowRestoreConfirmDialog_DoesNotCrash(t *testing.T) {
	fyneApp := test.NewApp()
	defer fyneApp.Quit()
	win := fyneApp.NewWindow("t")
	win.Show()
	defer win.Close()

	svc := &Services{Cfg: &config.Settings{}, Store: &stubStore{}}
	ShowRestoreConfirmDialog(svc,
		domain.Save{ID: "1", Name: "X", Path: "/tmp/x"},
		domain.Backup{ID: "b1", CreatedAt: time.Now(), Format: domain.BackupFormatZip},
		func() {},
	)
}
