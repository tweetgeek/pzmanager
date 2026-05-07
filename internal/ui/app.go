package ui

import (
	"context"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"pzm/internal/config"
	"pzm/internal/launcher"
	"pzm/internal/platform"
	"pzm/internal/storage"
)

// Services — zbiór zależności wstrzykiwanych do widoków UI.
type Services struct {
	Cfg      *config.Settings
	CfgPath  string
	Scanner  storage.SaveScanner
	Store    storage.BackupStore
	Platform platform.Platform
	Launcher *launcher.Launcher
}

// App — root widoku managera (okno główne z sidebarem).
type App struct {
	fyneApp  fyne.App
	window   fyne.Window
	services *Services
	content  *fyne.Container
	nav      *NavSidebar
}

func NewApp(fyneApp fyne.App, svc *Services) *App {
	a := &App{
		fyneApp:  fyneApp,
		services: svc,
		content:  container.NewStack(),
	}
	a.window = fyneApp.NewWindow("Project Zomboid Manager")
	a.window.Resize(fyne.NewSize(900, 600))

	a.nav = NewNavSidebar(a.onSelect)
	split := container.NewBorder(nil, nil, a.nav.Container(), nil, a.content)
	a.window.SetContent(split)

	a.showSaves() // ekran startowy
	return a
}

func (a *App) Show()       { a.window.Show() }
func (a *App) ShowAndRun() { a.window.ShowAndRun() }

func (a *App) onSelect(section string) {
	switch section {
	case "saves":
		a.showSaves()
	case "settings":
		a.showSettings()
	case "about":
		a.showAbout()
	case "launch":
		a.handleLaunch()
	}
}

func (a *App) swap(view fyne.CanvasObject) {
	a.content.Objects = []fyne.CanvasObject{view}
	a.content.Refresh()
}

// Stub'y — wypełniane w kolejnych Taskach.
func (a *App) showSaves()    { a.swap(NewSavesView(a.services, a.swap)) }
func (a *App) showSettings() { a.swap(NewSettingsView(a.services, a.swap)) }
func (a *App) showAbout() { a.swap(NewAboutView(a.services, a.swap)) }
func (a *App) handleLaunch() {
	if a.services.Launcher == nil {
		return
	}
	pw := NewProgressWindow(a.fyneApp)
	pw.Show()
	go func() {
		ctx := context.Background()
		if err := a.services.Launcher.AutoLaunch(ctx, nil, pw); err != nil {
			log.Printf("launch failed: %v", err)
		}
		pw.Close()
		if a.services.Cfg.CloseOnLaunch {
			fyne.Do(func() { a.window.Close() })
		}
	}()
}
