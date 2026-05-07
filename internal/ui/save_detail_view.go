package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"pzm/internal/domain"
)

func NewSaveDetailView(svc *Services, save domain.Save, swap func(fyne.CanvasObject)) fyne.CanvasObject {
	backups, _ := svc.Store.List(save.ID)

	breadcrumb := widget.NewButton("← Saves", func() {
		swap(NewSavesView(svc, swap))
	})
	title := widget.NewLabelWithStyle(save.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	header := container.NewBorder(nil, nil, breadcrumb, nil, title)

	list := widget.NewList(
		func() int { return len(backups) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			b := backups[i]
			icon := "📦"
			if b.Kind == domain.BackupKindManual {
				icon = "⭐"
			}
			label := b.Label
			if label != "" {
				label = " · " + label
			}
			o.(*widget.Label).SetText(fmt.Sprintf("%s %s · %s · %s · %d MB%s",
				icon,
				b.CreatedAt.Format("2006-01-02 15:04"),
				b.Kind.String(),
				b.Format.String(),
				b.SizeBytes/1024/1024,
				label,
			))
		},
	)

	var selected *domain.Backup
	list.OnSelected = func(i widget.ListItemID) { selected = &backups[i] }

	refresh := func() { swap(NewSaveDetailView(svc, save, swap)) }

	newBtn := widget.NewButton("+ New Backup", func() {
		ShowBackupCreateDialog(svc, save, refresh)
	})
	restoreBtn := widget.NewButton("⟲ Restore", func() {
		if selected == nil {
			return
		}
		ShowRestoreConfirmDialog(svc, save, *selected, refresh)
	})
	deleteBtn := widget.NewButton("🗑 Delete", func() {
		if selected == nil {
			return
		}
		win := currentWindow()
		if win == nil {
			return
		}
		sel := *selected
		dialog.ShowConfirm("Delete backup?",
			"This cannot be undone.",
			func(ok bool) {
				if !ok {
					return
				}
				if err := svc.Store.Delete(sel.ID); err != nil {
					dialog.ShowError(err, win)
					return
				}
				refresh()
			},
			win,
		)
	})

	bottom := container.NewHBox(newBtn, restoreBtn, deleteBtn)
	return container.NewBorder(header, bottom, nil, nil, list)
}

// currentWindow returns the first Fyne application window (safe because the manager has one).
// Used by dialogs.
func currentWindow() fyne.Window {
	app := fyne.CurrentApp()
	if app == nil {
		return nil
	}
	wins := app.Driver().AllWindows()
	if len(wins) == 0 {
		return nil
	}
	return wins[0]
}
