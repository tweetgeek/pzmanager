package domain

import "time"

type Platform int

const (
	PlatformWindows Platform = iota + 1
	PlatformLinux
	PlatformMacOS
)

func (p Platform) String() string {
	switch p {
	case PlatformWindows:
		return "windows"
	case PlatformLinux:
		return "linux"
	case PlatformMacOS:
		return "macos"
	default:
		return "unknown"
	}
}

type Source int

const (
	SourceUnknown Source = iota
	SourceSteam
	SourceGOG
	SourceManual
)

func (s Source) String() string {
	switch s {
	case SourceSteam:
		return "steam"
	case SourceGOG:
		return "gog"
	case SourceManual:
		return "manual"
	default:
		return "unknown"
	}
}

type GameInstance struct {
	Platform    Platform
	InstallPath string
	SavesPath   string
	Version     string
	Source      Source
	SteamAppID  string
}

type Save struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	GameMode    string    `json:"game_mode"`
	Path        string    `json:"path"`
	LastPlayed  time.Time `json:"last_played"`
	SizeBytes   int64     `json:"size_bytes"`
	GameVersion string    `json:"game_version,omitempty"`
}
