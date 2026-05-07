package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestNewAboutView_BuildsView(t *testing.T) {
	fyneApp := test.NewApp()
	defer fyneApp.Quit()

	view := NewAboutView(&Services{}, func(fyne.CanvasObject) {})
	if view == nil {
		t.Fatal("view nil")
	}
}
