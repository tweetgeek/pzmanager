package domain

import "testing"

func TestPlatform_String(t *testing.T) {
	cases := map[Platform]string{
		PlatformWindows: "windows",
		PlatformLinux:   "linux",
		PlatformMacOS:   "macos",
	}
	for p, want := range cases {
		if got := p.String(); got != want {
			t.Errorf("Platform(%d).String() = %q, want %q", p, got, want)
		}
	}
}

func TestSourceString(t *testing.T) {
	if SourceSteam.String() != "steam" {
		t.Errorf("got %q", SourceSteam.String())
	}
}

func TestSave_HasFields(t *testing.T) {
	s := Save{ID: "abc", Name: "Survivor_01", GameMode: "Apocalypse"}
	if s.ID != "abc" || s.Name != "Survivor_01" || s.GameMode != "Apocalypse" {
		t.Fatal("Save fields not assigned correctly")
	}
}
