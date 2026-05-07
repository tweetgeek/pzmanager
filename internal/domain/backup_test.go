package domain

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestBackupKind_String(t *testing.T) {
	if BackupKindAuto.String() != "auto" {
		t.Errorf("expected 'auto', got %q", BackupKindAuto.String())
	}
	if BackupKindManual.String() != "manual" {
		t.Errorf("expected 'manual', got %q", BackupKindManual.String())
	}
}

func TestBackupFormat_String(t *testing.T) {
	if BackupFormatZip.String() != "zip" {
		t.Errorf("expected 'zip', got %q", BackupFormatZip.String())
	}
	if BackupFormatFolder.String() != "folder" {
		t.Errorf("expected 'folder', got %q", BackupFormatFolder.String())
	}
}

func TestParseBackupFormat(t *testing.T) {
	cases := []struct {
		in      string
		want    BackupFormat
		wantErr bool
	}{
		{"zip", BackupFormatZip, false},
		{"folder", BackupFormatFolder, false},
		{"ZIP", BackupFormatZip, false},
		{"unknown", 0, true},
	}
	for _, c := range cases {
		got, err := ParseBackupFormat(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("ParseBackupFormat(%q) err = %v, wantErr=%v", c.in, err, c.wantErr)
		}
		if !c.wantErr && got != c.want {
			t.Errorf("ParseBackupFormat(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestBackup_JSON_RoundTrip_PreservesEnumStrings(t *testing.T) {
	orig := Backup{
		ID:        "abc",
		SaveID:    "save1",
		Kind:      BackupKindAuto,
		Format:    BackupFormatZip,
		Path:      "/x",
		CreatedAt: time.Date(2026, 5, 6, 14, 22, 0, 0, time.UTC),
	}
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatal(err)
	}
	// human-readable enum values:
	s := string(data)
	if !strings.Contains(s, `"kind":"auto"`) {
		t.Errorf("expected kind:\"auto\" in JSON, got: %s", s)
	}
	if !strings.Contains(s, `"format":"zip"`) {
		t.Errorf("expected format:\"zip\" in JSON, got: %s", s)
	}

	var back Backup
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if back.Kind != BackupKindAuto {
		t.Errorf("Kind round-trip failed: %v", back.Kind)
	}
	if back.Format != BackupFormatZip {
		t.Errorf("Format round-trip failed: %v", back.Format)
	}
}

func TestParseBackupKind(t *testing.T) {
	cases := []struct {
		in      string
		want    BackupKind
		wantErr bool
	}{
		{"auto", BackupKindAuto, false},
		{"manual", BackupKindManual, false},
		{"AUTO", BackupKindAuto, false},
		{"bogus", 0, true},
	}
	for _, c := range cases {
		got, err := ParseBackupKind(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("ParseBackupKind(%q) err=%v wantErr=%v", c.in, err, c.wantErr)
		}
		if !c.wantErr && got != c.want {
			t.Errorf("ParseBackupKind(%q)=%v want %v", c.in, got, c.want)
		}
	}
}
