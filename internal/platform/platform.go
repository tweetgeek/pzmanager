package platform

import "pzm/internal/domain"

// Platform abstrahuje OS-specific operacje: ścieżki domyślne,
// detekcję gry, uruchamianie procesów, instalację skrótów.
//
// Konkretna implementacja jest wybierana w czasie kompilacji przez build
// tagi w plikach linux.go / darwin.go / windows.go — każdy z nich definiuje
// swoją wersję `New()` widoczną tylko dla danego GOOS.
type Platform interface {
	Identity() domain.Platform
	DefaultSavesPath() string
	DefaultGameInstallPath() string
	DefaultBackupPath() string
	DefaultConfigPath() string
	DetectGame() (domain.GameInstance, error)
	LaunchGame(inst domain.GameInstance, steamCommand []string) error
	InstallShortcut(target string) error
}
