package ui

import (
	"errors"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func NewSettingsView(svc *Services, swap func(fyne.CanvasObject)) fyne.CanvasObject {
	win := currentWindow()

	savesEntry := widget.NewEntry()
	savesEntry.SetText(svc.Cfg.SavesPath)
	installEntry := widget.NewEntry()
	installEntry.SetText(svc.Cfg.GameInstallPath)
	backupEntry := widget.NewEntry()
	backupEntry.SetText(svc.Cfg.BackupRoot)

	autoBackup := widget.NewCheck("Enable auto-backup before launch", nil)
	autoBackup.SetChecked(svc.Cfg.AutoBackup)

	retainEntry := widget.NewEntry()
	retainEntry.SetText(strconv.Itoa(svc.Cfg.AutoBackupRetain))

	formatSel := widget.NewRadioGroup([]string{"ZIP", "Folder"}, nil)
	if svc.Cfg.DefaultFormat == "folder" {
		formatSel.SetSelected("Folder")
	} else {
		formatSel.SetSelected("ZIP")
	}

	closeAfterLaunch := widget.NewCheck("Close manager after launching game", nil)
	closeAfterLaunch.SetChecked(svc.Cfg.CloseOnLaunch)

	saveBtn := widget.NewButton("Save", func() {
		if !pathExists(savesEntry.Text) {
			if win != nil {
				dialog.ShowError(errors.New("Saves path does not exist: "+savesEntry.Text), win)
			}
			return
		}
		retain, err := strconv.Atoi(retainEntry.Text)
		if err != nil || retain < 1 {
			if win != nil {
				dialog.ShowError(errors.New("Retention must be a positive integer"), win)
			}
			return
		}

		svc.Cfg.SavesPath = savesEntry.Text
		svc.Cfg.GameInstallPath = installEntry.Text
		svc.Cfg.BackupRoot = backupEntry.Text
		svc.Cfg.AutoBackup = autoBackup.Checked
		svc.Cfg.AutoBackupRetain = retain
		svc.Cfg.CloseOnLaunch = closeAfterLaunch.Checked
		if formatSel.Selected == "Folder" {
			svc.Cfg.DefaultFormat = "folder"
		} else {
			svc.Cfg.DefaultFormat = "zip"
		}
		if err := svc.Cfg.Save(svc.CfgPath); err != nil {
			if win != nil {
				dialog.ShowError(err, win)
			}
			return
		}
		if win != nil {
			dialog.ShowInformation("Saved", "Settings saved.", win)
		}
	})

	autoDetectBtn := widget.NewButton("Auto-detect", func() {
		if svc.Platform == nil {
			return
		}
		inst, err := svc.Platform.DetectGame()
		if err == nil {
			savesEntry.SetText(inst.SavesPath)
			installEntry.SetText(inst.InstallPath)
		} else {
			savesEntry.SetText(svc.Platform.DefaultSavesPath())
			installEntry.SetText(svc.Platform.DefaultGameInstallPath())
			if backupEntry.Text == "" {
				backupEntry.SetText(svc.Platform.DefaultBackupPath())
			}
		}
	})

	launcherBtn := widget.NewButton("Launcher Setup...", func() {
		ShowLauncherWizard(svc)
	})

	paths := widget.NewCard("Paths", "Where to find game files",
		container.NewVBox(
			widget.NewLabel("Saves:"), savesEntry,
			widget.NewLabel("Game install:"), installEntry,
			widget.NewLabel("Backup directory:"), backupEntry,
			autoDetectBtn,
		),
	)
	auto := widget.NewCard("Auto-backup", "Behavior when launching the game",
		container.NewVBox(
			autoBackup,
			widget.NewLabel("Retain N latest auto-backups:"), retainEntry,
			widget.NewLabel("Default format:"), formatSel,
			closeAfterLaunch,
		),
	)
	launcher := widget.NewCard("Launcher", "Configure how the game is started",
		launcherBtn,
	)

	return container.NewVScroll(container.NewVBox(paths, auto, launcher, saveBtn))
}

func pathExists(p string) bool {
	if p == "" {
		return false
	}
	_, err := os.Stat(p)
	return err == nil
}
