package launcher

import (
	"context"
	"errors"
	"testing"
	"time"

	"pzm/internal/config"
	"pzm/internal/domain"
	"pzm/internal/storage"
)

// fake implementations of storage.SaveScanner / storage.BackupStore / launcher.GamePlatform.

type fakeScanner struct {
	saves []domain.Save
	hash  string
	err   error
}

func (f *fakeScanner) Scan(string) ([]domain.Save, error) { return f.saves, f.err }
func (f *fakeScanner) Hash(domain.Save) (string, error)   { return f.hash, nil }

type fakeStore struct {
	backups      []domain.Backup
	created      []domain.Backup
	pruned       int
	failOnCreate bool
}

func (f *fakeStore) List(string) ([]domain.Backup, error) { return f.backups, nil }
func (f *fakeStore) Create(s domain.Save, opts storage.CreateOpts) (domain.Backup, error) {
	if f.failOnCreate {
		return domain.Backup{}, errors.New("disk full")
	}
	b := domain.Backup{
		ID: "newbackup", SaveID: s.ID, Kind: opts.Kind, Format: opts.Format,
		CreatedAt: time.Now(), SaveHash: "fresh",
	}
	f.created = append(f.created, b)
	return b, nil
}
func (f *fakeStore) Restore(domain.Backup, string) error { return nil }
func (f *fakeStore) Delete(string) error                  { return nil }
func (f *fakeStore) PruneAuto(string, int) (int, error)   { f.pruned++; return 1, nil }

type fakePlatform struct {
	launched bool
	cmd      []string
}

func (f *fakePlatform) LaunchGame(inst domain.GameInstance, c []string) error {
	f.launched = true
	f.cmd = c
	return nil
}
func (f *fakePlatform) DetectGame() (domain.GameInstance, error) {
	return domain.GameInstance{}, nil
}

func TestAutoLaunch_NoSavesChanged_SkipsBackup(t *testing.T) {
	sc := &fakeScanner{
		saves: []domain.Save{{ID: "save1", Name: "S", LastPlayed: time.Now()}},
		hash:  "same-hash",
	}
	st := &fakeStore{
		backups: []domain.Backup{{
			ID: "old", SaveID: "save1", Kind: domain.BackupKindAuto,
			SaveHash: "same-hash", CreatedAt: time.Now().Add(-time.Hour),
		}},
	}
	pl := &fakePlatform{}
	cfg := &config.Settings{AutoBackup: true, AutoBackupRetain: 5, DefaultFormat: "zip"}

	l := New(sc, st, pl, cfg)
	err := l.AutoLaunch(context.Background(), nil, NewNoopReporter())
	if err != nil {
		t.Fatal(err)
	}
	if len(st.created) != 0 {
		t.Errorf("expected backup skipped, got %d created", len(st.created))
	}
	if !pl.launched {
		t.Error("expected game launched")
	}
}

func TestAutoLaunch_ChangedSave_CreatesBackupAndPrunes(t *testing.T) {
	sc := &fakeScanner{
		saves: []domain.Save{{ID: "save1", Name: "S", LastPlayed: time.Now()}},
		hash:  "fresh",
	}
	st := &fakeStore{
		backups: []domain.Backup{{
			ID: "old", SaveID: "save1", Kind: domain.BackupKindAuto,
			SaveHash: "stale", CreatedAt: time.Now().Add(-time.Hour),
		}},
	}
	pl := &fakePlatform{}
	cfg := &config.Settings{AutoBackup: true, AutoBackupRetain: 3, DefaultFormat: "zip"}

	l := New(sc, st, pl, cfg)
	err := l.AutoLaunch(context.Background(), []string{"steam-launch"}, NewNoopReporter())
	if err != nil {
		t.Fatal(err)
	}
	if len(st.created) != 1 {
		t.Errorf("expected 1 backup, got %d", len(st.created))
	}
	if st.pruned != 1 {
		t.Errorf("expected PruneAuto called once, got %d", st.pruned)
	}
	if !pl.launched {
		t.Error("expected game launched")
	}
	if len(pl.cmd) != 1 || pl.cmd[0] != "steam-launch" {
		t.Errorf("expected steam command passed, got %v", pl.cmd)
	}
}

func TestAutoLaunch_BackupFails_StillLaunchesGame(t *testing.T) {
	sc := &fakeScanner{
		saves: []domain.Save{{ID: "save1", Name: "S", LastPlayed: time.Now()}},
		hash:  "fresh",
	}
	st := &fakeStore{failOnCreate: true}
	pl := &fakePlatform{}
	cfg := &config.Settings{AutoBackup: true, AutoBackupRetain: 5, DefaultFormat: "zip"}

	l := New(sc, st, pl, cfg)
	err := l.AutoLaunch(context.Background(), nil, NewNoopReporter())
	if err != nil {
		t.Fatalf("expected nil err (game still launches), got %v", err)
	}
	if !pl.launched {
		t.Error("game should launch even when backup fails")
	}
}

func TestAutoLaunch_AutoBackupDisabled_SkipsBackup(t *testing.T) {
	sc := &fakeScanner{
		saves: []domain.Save{{ID: "save1", LastPlayed: time.Now()}},
		hash:  "anything",
	}
	st := &fakeStore{}
	pl := &fakePlatform{}
	cfg := &config.Settings{AutoBackup: false, DefaultFormat: "zip"}

	l := New(sc, st, pl, cfg)
	if err := l.AutoLaunch(context.Background(), nil, NewNoopReporter()); err != nil {
		t.Fatal(err)
	}
	if len(st.created) != 0 {
		t.Errorf("expected no backup, got %d", len(st.created))
	}
	if !pl.launched {
		t.Error("expected game launched")
	}
}
