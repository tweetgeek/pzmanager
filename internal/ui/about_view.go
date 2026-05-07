package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Version jest ustawiana przez ldflags z cmd/manager/version.go.
// W UI używamy lokalnej kopii — w produkcji wstrzykujemy przez main.
var Version = "dev"

func NewAboutView(svc *Services, swap func(fyne.CanvasObject)) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Project Zomboid Manager",
		fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	versionLbl := widget.NewLabel("Version: " + Version)
	desc := widget.NewLabel("Cross-platform launcher and save backup manager.")
	return container.NewCenter(container.NewVBox(title, versionLbl, desc))
}
