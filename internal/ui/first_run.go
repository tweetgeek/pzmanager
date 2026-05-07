package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowFirstRunWizard pokazywany gdy settings.json nie istniał (first launch).
// Auto-detekcja ścieżek + zapis. Pokazuje user'owi co zostało wykryte.
func ShowFirstRunWizard(svc *Services, onDone func()) {
	win := currentWindow()
	if win == nil {
		if onDone != nil {
			onDone()
		}
		return
	}

	if svc.Platform != nil {
		inst, err := svc.Platform.DetectGame()
		if err == nil {
			svc.Cfg.SavesPath = inst.SavesPath
			svc.Cfg.GameInstallPath = inst.InstallPath
		} else {
			svc.Cfg.SavesPath = svc.Platform.DefaultSavesPath()
			svc.Cfg.GameInstallPath = svc.Platform.DefaultGameInstallPath()
		}
		svc.Cfg.BackupRoot = svc.Platform.DefaultBackupPath()
	}

	body := widget.NewLabel(
		"Welcome to Project Zomboid Manager.\n\n" +
			"Detected paths:\n" +
			"  Saves: " + svc.Cfg.SavesPath + "\n" +
			"  Game: " + svc.Cfg.GameInstallPath + "\n" +
			"  Backups: " + svc.Cfg.BackupRoot + "\n\n" +
			"You can change these any time in Settings.",
	)
	body.Wrapping = fyne.TextWrapWord

	d := dialog.NewCustomConfirm(
		"Welcome",
		"Get Started",
		"Skip",
		container.NewVBox(body),
		func(ok bool) {
			_ = svc.Cfg.Save(svc.CfgPath)
			if onDone != nil {
				onDone()
			}
		},
		win,
	)
	d.Resize(fyne.NewSize(500, 350))
	d.Show()
}
