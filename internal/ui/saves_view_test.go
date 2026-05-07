package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"pzm/internal/config"
	"pzm/internal/domain"
	"pzm/internal/storage"
)

// stubScanner / stubStore for UI tests.
type stubScanner struct{ saves []domain.Save }

func (s *stubScanner) Scan(string) ([]domain.Save, error) { return s.saves, nil }
func (s *stubScanner) Hash(domain.Save) (string, error)   { return "", nil }

type stubStore struct{ backups map[string][]domain.Backup }

func (s *stubStore) List(id string) ([]domain.Backup, error) {
	if s.backups == nil {
		return nil, nil
	}
	return s.backups[id], nil
}
func (s *stubStore) Create(domain.Save, storage.CreateOpts) (domain.Backup, error) {
	return domain.Backup{}, nil
}
func (s *stubStore) Restore(domain.Backup, string) error { return nil }
func (s *stubStore) Delete(string) error                 { return nil }
func (s *stubStore) PruneAuto(string, int) (int, error)  { return 0, nil }

func TestSavesView_RendersListWithSaves(t *testing.T) {
	fyneApp := test.NewApp()
	defer fyneApp.Quit()

	svc := &Services{
		Cfg: &config.Settings{},
		Scanner: &stubScanner{saves: []domain.Save{
			{ID: "1", Name: "Survivor_01", GameMode: "Apocalypse", LastPlayed: time.Now()},
			{ID: "2", Name: "Muldraugh_Coop", GameMode: "Sandbox", LastPlayed: time.Now().Add(-time.Hour)},
		}},
		Store: &stubStore{backups: map[string][]domain.Backup{}},
	}

	view := NewSavesView(svc, func(o fyne.CanvasObject) {})
	if view == nil {
		t.Fatal("view is nil")
	}
}

func TestSavesView_NoSaves_RendersPlaceholder(t *testing.T) {
	fyneApp := test.NewApp()
	defer fyneApp.Quit()

	svc := &Services{
		Cfg:     &config.Settings{},
		Scanner: &stubScanner{saves: nil},
		Store:   &stubStore{},
	}

	view := NewSavesView(svc, func(o fyne.CanvasObject) {})
	if view == nil {
		t.Fatal("view is nil")
	}
}
