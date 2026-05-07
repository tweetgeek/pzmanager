package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func ShowLauncherWizard(svc *Services) {
	win := currentWindow()
	if win == nil {
		return
	}

	explanation := widget.NewLabel(
		"Three ways to launch the game with auto-backup:\n\n" +
			"(A) From this app — click 'Launch Game' in the sidebar.\n\n" +
			"(B) Desktop shortcut — adds a shortcut that calls this manager.\n\n" +
			"(C) Steam Launch Options — copies a string to paste in Steam → " +
			"Properties → Launch Options. This intercepts every Steam launch.",
	)
	explanation.Wrapping = fyne.TextWrapWord

	installShortcutBtn := widget.NewButton("Install desktop shortcut", func() {
		exe := executablePath()
		if exe == "" || svc.Platform == nil {
			return
		}
		if err := svc.Platform.InstallShortcut(exe); err != nil {
			dialog.ShowError(err, win)
			return
		}
		svc.Cfg.LauncherInstalled.Shortcut = true
		_ = svc.Cfg.Save(svc.CfgPath)
		dialog.ShowInformation("Done", "Shortcut installed on desktop.", win)
	})

	copySteamBtn := widget.NewButton("Copy Steam Launch Options", func() {
		exe := executablePath()
		if exe == "" {
			return
		}
		cmd := `"` + exe + `" --auto-launch -- %command%`
		win.Clipboard().SetContent(cmd)
		svc.Cfg.LauncherInstalled.SteamCopied = true
		_ = svc.Cfg.Save(svc.CfgPath)
		dialog.ShowInformation("Copied", "Paste this in Steam → Properties → Launch Options:\n\n"+cmd, win)
	})

	body := container.NewVBox(explanation, installShortcutBtn, copySteamBtn)
	d := dialog.NewCustom("Launcher Setup", "Close", body, win)
	d.Resize(fyne.NewSize(500, 400))
	d.Show()
}
