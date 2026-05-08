package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Version jest ustawiana przez ldflags z cmd/manager/version.go.
// W UI używamy lokalnej kopii — w produkcji wstrzykujemy przez main.
var Version = "dev"

func NewAboutView(svc *Services, swap func(fyne.CanvasObject)) fyne.CanvasObject {
	icon := canvas.NewImageFromResource(AppIcon)
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(192, 192))

	title := widget.NewLabelWithStyle("Project Zomboid Manager",
		fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	versionLbl := widget.NewLabelWithStyle("Version: "+Version,
		fyne.TextAlignCenter, fyne.TextStyle{})
	desc := widget.NewLabelWithStyle("Cross-platform launcher and save backup manager.",
		fyne.TextAlignCenter, fyne.TextStyle{})

	return container.NewCenter(container.NewVBox(
		container.NewCenter(icon),
		title,
		versionLbl,
		desc,
	))
}
