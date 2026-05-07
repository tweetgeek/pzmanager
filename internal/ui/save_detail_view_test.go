package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"pzm/internal/config"
	"pzm/internal/domain"
)

func TestSaveDetailView_BuildsWithBackups(t *testing.T) {
	fyneApp := test.NewApp()
	defer fyneApp.Quit()

	save := domain.Save{ID: "s1", Name: "Test", GameMode: "Apoc"}
	svc := &Services{
		Cfg: &config.Settings{},
		Store: &stubStore{
			backups: map[string][]domain.Backup{
				"s1": {
					{ID: "b1", SaveID: "s1", Kind: domain.BackupKindManual, CreatedAt: time.Now(), Label: "before raid", Format: domain.BackupFormatZip},
					{ID: "b2", SaveID: "s1", Kind: domain.BackupKindAuto, CreatedAt: time.Now().Add(-time.Hour), Format: domain.BackupFormatFolder},
				},
			},
		},
	}
	view := NewSaveDetailView(svc, save, func(o fyne.CanvasObject) {})
	if view == nil {
		t.Fatal("view is nil")
	}
}

func TestSaveDetailView_NoBackups_BuildsEmpty(t *testing.T) {
	fyneApp := test.NewApp()
	defer fyneApp.Quit()

	save := domain.Save{ID: "s2", Name: "Empty", GameMode: "Apoc"}
	svc := &Services{
		Cfg:   &config.Settings{},
		Store: &stubStore{backups: nil},
	}
	view := NewSaveDetailView(svc, save, func(o fyne.CanvasObject) {})
	if view == nil {
		t.Fatal("view is nil")
	}
}
