package main

import (
	"embed"
	"fmt"
	"log"
	"time"

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

	// Connect to database with retry
	var database *db.DB
	var err error
	for i := 1; i <= 3; i++ {
		log.Printf("Connecting to database (Attempt %d/3): %s", i, cfg.DBConnectionString)
		database, err = db.Connect(cfg)
		if err == nil {
			break
		}
		log.Printf("Connection attempt %d failed: %v", i, err)
		if i < 3 {
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		log.Fatalf("Could not connect to database after 3 attempts: %v", err)
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
