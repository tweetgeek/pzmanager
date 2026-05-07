package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"pzm/internal/domain"
	"pzm/internal/storage"
)

func ShowRestoreConfirmDialog(svc *Services, save domain.Save, backup domain.Backup, onDone func()) {
	win := currentWindow()
	if win == nil {
		return
	}

	info := widget.NewLabel(fmt.Sprintf("Restore backup from %s for save %s?",
		backup.CreatedAt.Format("2006-01-02 15:04"), save.Name))
	if backup.Label != "" {
		info.SetText(info.Text + "\nLabel: " + backup.Label)
	}
	safetyCheck := widget.NewCheck("Back up current save before restoring", nil)
	safetyCheck.SetChecked(true)

	body := container.NewVBox(info, safetyCheck)

	var d dialog.Dialog
	d = dialog.NewCustomConfirm(
		"Restore Backup",
		"Restore",
		"Cancel",
		body,
		func(ok bool) {
			if !ok {
				return
			}
			doSafety := safetyCheck.Checked
			go func() {
				if doSafety {
					if _, err := svc.Store.Create(save, storage.CreateOpts{
						Format: domain.BackupFormatZip,
						Kind:   domain.BackupKindAuto,
						Label:  "pre-restore safety",
					}); err != nil {
						fyne.Do(func() { dialog.ShowError(fmt.Errorf("safety backup failed: %w", err), win) })
						return
					}
				}
				if err := svc.Store.Restore(backup, save.Path); err != nil {
					fyne.Do(func() { dialog.ShowError(err, win) })
					return
				}
				fyne.Do(func() {
					if onDone != nil {
						onDone()
					}
				})
			}()
		},
		win,
	)
	d.Resize(fyne.NewSize(400, 200))
	d.Show()
}
