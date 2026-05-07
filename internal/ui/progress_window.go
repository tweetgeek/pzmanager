package ui

import (
	"context"
	"log"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"pzm/internal/launcher"
)

// ProgressWindow — UI dla auto-launch (CLI mode + GUI launch button).
// Implementuje launcher.ProgressReporter.
type ProgressWindow struct {
	win      fyne.Window
	status   *widget.Label
	progress *widget.ProgressBar
	skipBtn  *widget.Button
	skipCh   chan struct{}
	skipOnce sync.Once
}

func NewProgressWindow(fyneApp fyne.App) *ProgressWindow {
	p := &ProgressWindow{
		win:      fyneApp.NewWindow("Launching..."),
		status:   widget.NewLabel("Starting..."),
		progress: widget.NewProgressBar(),
		skipCh:   make(chan struct{}),
	}
	p.skipBtn = widget.NewButton("Skip backup, launch now", func() {
		p.skipOnce.Do(func() { close(p.skipCh) })
		p.skipBtn.Disable()
	})

	body := container.NewVBox(p.status, p.progress, p.skipBtn)
	p.win.SetContent(body)
	p.win.Resize(fyne.NewSize(400, 180))
	p.win.SetFixedSize(true)
	return p
}

func (p *ProgressWindow) Show() { p.win.Show() }
func (p *ProgressWindow) Close() {
	fyne.Do(func() { p.win.Close() })
}

// ProgressReporter implementation
func (p *ProgressWindow) Status(msg string) {
	fyne.Do(func() { p.status.SetText(msg) })
}
func (p *ProgressWindow) Progress(f float64) {
	fyne.Do(func() { p.progress.SetValue(f) })
}
func (p *ProgressWindow) OnSkip() <-chan struct{} { return p.skipCh }

// RunAutoLaunch wykonuje pełen flow auto-launch z UI feedbackiem.
// Po zakończeniu (sukces/błąd) okno się zamyka i Fyne app kończy.
// Używane w trybie CLI gdy aplikacja startuje z --auto-launch.
func RunAutoLaunch(fyneApp fyne.App, l *launcher.Launcher, steamCmd []string) {
	pw := NewProgressWindow(fyneApp)
	pw.Show()

	go func() {
		ctx := context.Background()
		if err := l.AutoLaunch(ctx, steamCmd, pw); err != nil {
			log.Printf("auto-launch error: %v", err)
		}
		pw.Close()
		fyneApp.Quit()
	}()
}
