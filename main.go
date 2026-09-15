package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"faster/internal/db"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	dbPath, err := db.DefaultPath()
	if err != nil {
		log.Fatalf("resolve database path: %v", err)
	}

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("open database at %s: %v", dbPath, err)
	}
	defer database.Close()

	app := NewApp(database)

	// Dark, near-monochromatic background so the native window frame
	// doesn't flash white before the frontend paints its own background.
	err = wails.Run(&options.App{
		Title:  "Faster",
		Width:  980,
		Height: 720,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 18, B: 20, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatalf("run app: %v", err)
	}
}
