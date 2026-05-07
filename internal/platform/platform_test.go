package platform

import (
	"runtime"
	"testing"

	"pzm/internal/domain"
)

func TestNew_ReturnsCorrectPlatformForGoos(t *testing.T) {
	p := New()
	var want domain.Platform
	switch runtime.GOOS {
	case "windows":
		want = domain.PlatformWindows
	case "linux":
		want = domain.PlatformLinux
	case "darwin":
		want = domain.PlatformMacOS
	default:
		t.Skipf("unsupported GOOS: %s", runtime.GOOS)
	}
	if p.Identity() != want {
		t.Errorf("Identity() = %v, want %v", p.Identity(), want)
	}
}
