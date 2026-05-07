package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func NewSavesView(svc *Services, swap func(fyne.CanvasObject)) fyne.CanvasObject {
	if svc.Scanner == nil {
		return placeholderNoSaves(svc, swap)
	}
	saves, err := svc.Scanner.Scan(svc.Cfg.SavesPath)
	if err != nil || len(saves) == 0 {
		return placeholderNoSaves(svc, swap)
	}

	title := widget.NewLabelWithStyle("Saves", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	list := widget.NewList(
		func() int { return len(saves) },
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabel("Name"),
				widget.NewLabel("meta"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			row := o.(*fyne.Container).Objects
			row[0].(*widget.Label).SetText(saves[i].Name)

			backups, _ := svc.Store.List(saves[i].ID)
			meta := fmt.Sprintf("%s · %s · last played: %s · backups: %d",
				saves[i].GameMode,
				ifEmpty(saves[i].GameVersion, "v?"),
				saves[i].LastPlayed.Format("2006-01-02 15:04"),
				len(backups))
			row[1].(*widget.Label).SetText(meta)
		},
	)
	list.OnSelected = func(i widget.ListItemID) {
		swap(NewSaveDetailView(svc, saves[i], swap))
	}

	refreshBtn := widget.NewButton("Refresh", func() {
		swap(NewSavesView(svc, swap))
	})

	top := container.NewBorder(nil, nil, title, refreshBtn)
	return container.NewBorder(top, nil, nil, nil, list)
}

func placeholderNoSaves(svc *Services, swap func(fyne.CanvasObject)) fyne.CanvasObject {
	msg := widget.NewLabel("No saves found.")
	btn := widget.NewButton("Open Settings", func() {
		swap(NewSettingsView(svc, swap))
	})
	return container.NewVBox(msg, btn)
}

func ifEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
