package main

import (
	"log"

	"fyne.io/fyne/v2/app"

	"multiout/internal/audio"
	"multiout/internal/config"
	"multiout/internal/engine"
	"multiout/internal/ui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("warn: failed to load config: %v", err)
	}

	be, err := audio.NewMalgoBackend()
	if err != nil {
		log.Fatalf("audio backend init failed: %v", err)
	}
	defer be.Close()

	eng := engine.NewEngine(be, audio.Config{SampleRate: 48000, Channels: 2})
	defer eng.Close()

	sinkMgr := audio.NewVirtualSinkManager()
	// Clean up any leftover virtual sink from past crashes
	_ = sinkMgr.CleanupStale()
	defer func() {
		_ = sinkMgr.Restore()
	}()

	fyneApp := app.NewWithID("com.codesbysagar.multiout")
	fyneApp.SetIcon(ui.LogoResource)
	fyneApp.Settings().SetTheme(ui.NewMultioutTheme(ui.ThemeMode(cfg.GetThemeMode())))

	ctrl := ui.NewController(be, eng, sinkMgr, cfg)

	mainWindow := ui.NewMainWindow(fyneApp, ctrl)
	mainWindow.ShowAndRun()
}
