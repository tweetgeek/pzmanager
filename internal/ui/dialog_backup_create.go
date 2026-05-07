package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"pzm/internal/domain"
	"pzm/internal/storage"
)

func ShowBackupCreateDialog(svc *Services, save domain.Save, onDone func()) {
	win := currentWindow()
	if win == nil {
		return
	}

	formatSel := widget.NewRadioGroup([]string{"ZIP", "Folder"}, nil)
	if svc.Cfg.DefaultFormat == "folder" {
		formatSel.SetSelected("Folder")
	} else {
		formatSel.SetSelected("ZIP")
	}

	labelInput := widget.NewEntry()
	labelInput.SetPlaceHolder("Optional label, e.g. before raid")

	progress := widget.NewProgressBarInfinite()
	progress.Hide()

	form := container.NewVBox(
		widget.NewLabel("Format:"),
		formatSel,
		widget.NewLabel("Label:"),
		labelInput,
		progress,
	)

	var d dialog.Dialog
	d = dialog.NewCustomConfirm(
		"Create Backup",
		"Create Backup",
		"Cancel",
		form,
		func(ok bool) {
			if !ok {
				return
			}
			format := domain.BackupFormatZip
			if formatSel.Selected == "Folder" {
				format = domain.BackupFormatFolder
			}

			progress.Show()
			d.Hide()

			go func() {
				_, err := svc.Store.Create(save, storage.CreateOpts{
					Format: format,
					Kind:   domain.BackupKindManual,
					Label:  labelInput.Text,
				})
				fyne.Do(func() {
					if err != nil {
						dialog.ShowError(err, win)
						return
					}
					if onDone != nil {
						onDone()
					}
				})
			}()
		},
		win,
	)
	d.Resize(fyne.NewSize(400, 250))
	d.Show()
}
