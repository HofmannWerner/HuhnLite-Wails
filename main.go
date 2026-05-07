package main

import (
	"embed"
	"fmt"
	"log"

	"huhnlite-wails/backend/config"
	"huhnlite-wails/backend/db"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Load config
	cfg := config.LoadConfig()

	// Connect to database
	log.Printf("Connecting to database: %s (Engine: %s)", cfg.DBConnectionString, cfg.DBEngine)
	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	// Create an instance of the app structure
	app := NewApp(database)

	// Fenstergröße aus DB laden
	width, height := 1280, 800
	var savedSize string
	err = database.SQL.QueryRow("SELECT VALUE FROM USER_STATE WHERE KEY = 'window_size' ORDER BY USERNAME DESC LIMIT 1").Scan(&savedSize)
	if err == nil && savedSize != "" {
		fmt.Sscanf(savedSize, "%dx%d", &width, &height)
		log.Printf("Restoring window size: %dx%d", width, height)
	}

	// Maximiert-Status laden
	startState := options.Maximised
	var savedMax string
	err = database.SQL.QueryRow("SELECT VALUE FROM USER_STATE WHERE KEY = 'window_maximized' ORDER BY USERNAME DESC LIMIT 1").Scan(&savedMax)
	if err == nil {
		if savedMax == "false" {
			startState = options.Normal
		}
	}

	// Create application with options
	err = wails.Run(&options.App{
		Title:            "HuhnLite-Wails",
		Width:            width,
		Height:           height,
		WindowStartState: startState,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnBeforeClose:    app.beforeClose,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
