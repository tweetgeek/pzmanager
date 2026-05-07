package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestProgressWindow_BuildsAndImplementsReporter(t *testing.T) {
	fyneApp := test.NewApp()
	defer fyneApp.Quit()

	pw := NewProgressWindow(fyneApp)
	pw.Status("hello")
	pw.Progress(0.5)
	select {
	case <-pw.OnSkip():
		t.Error("expected skip channel open")
	default:
	}
}
