package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2/app"

	"pzm/internal/config"
	"pzm/internal/launcher"
	"pzm/internal/platform"
	"pzm/internal/storage"
	"pzm/internal/ui"
)

func main() {
	var (
		autoLaunch  bool
		showVersion bool
		verbose     bool
	)
	flag.BoolVar(&autoLaunch, "auto-launch", false, "run backup + launch game then exit")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.BoolVar(&verbose, "verbose", false, "verbose logging")
	flag.Parse()

	if showVersion {
		fmt.Println(Version)
		return
	}

	pl := platform.New()

	cfgPath := filepath.Join(pl.DefaultConfigPath(), "settings.json")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load settings: %v\n", err)
		os.Exit(1)
	}
	if cfg.SavesPath == "" {
		cfg.SavesPath = pl.DefaultSavesPath()
	}
	if cfg.BackupRoot == "" {
		cfg.BackupRoot = pl.DefaultBackupPath()
	}
	if cfg.GameInstallPath == "" {
		cfg.GameInstallPath = pl.DefaultGameInstallPath()
	}

	scanner := storage.NewFsSaveScanner()
	store := storage.NewFsBackupStore(cfg.BackupRoot)
	l := launcher.New(scanner, store, pl, cfg)

	fyneApp := app.New()
	ui.Version = Version

	services := &ui.Services{
		Cfg:      cfg,
		CfgPath:  cfgPath,
		Scanner:  scanner,
		Store:    store,
		Platform: pl,
		Launcher: l,
	}

	if autoLaunch {
		steamCmd := flag.Args()
		ui.RunAutoLaunch(fyneApp, l, steamCmd)
		fyneApp.Run()
		return
	}

	a := ui.NewApp(fyneApp, services)
	a.ShowAndRun()
}
