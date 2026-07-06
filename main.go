package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"huhnlite-wails/backend/api"
	"huhnlite-wails/backend/config"
	"huhnlite-wails/backend/db"

	"github.com/gin-gonic/gin"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Load config
	cfg := config.LoadConfig()

	if cfg.Test == 1 && cfg.DBConnectTest != "" {
		log.Printf("Test-Datenbank-Modus ist aktiv. Schalte um auf: %s", cfg.DBConnectTest)
		cfg.DBConnectionString = cfg.DBConnectTest
	}

	// Server-Modus: Starte reinen Gin Webserver und beende danach (ohne Wails-GUI)
	if cfg.Mode == "server" {
		log.Printf("Running in SERVER mode")

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
			log.Fatalf("Failed to connect to database: %v", err)
		}

		engine := api.StartServer(database)

		// Serve static frontend files from embed.FS
		subFS, err := fs.Sub(assets, "frontend/dist/spa")
		if err == nil {
			fileServer := http.FileServer(http.FS(subFS))
			engine.NoRoute(func(c *gin.Context) {
				path := c.Request.URL.Path
				if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/help") {
					return
				}
				// Prüfen, ob die Datei im embedded Dateisystem existiert
				file, err := subFS.Open(strings.TrimPrefix(path, "/"))
				if err == nil {
					file.Close()
					fileServer.ServeHTTP(c.Writer, c.Request)
					return
				}
				// Fallback auf index.html für SPA-Routing
				c.Request.URL.Path = "/"
				fileServer.ServeHTTP(c.Writer, c.Request)
			})
		} else {
			log.Printf("Failed to create sub FS for frontend: %v", err)
		}

		port := 8080
		if cfg.Port > 0 {
			port = cfg.Port
		}
		log.Printf("Starting HuhnLite Server on port %d", port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), engine); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
		return
	}

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

	// Create an instance of the app structure
	app := NewApp(database)
	if err != nil {
		app.ConnectionError = err.Error()
	}

	// Fenstergröße aus DB laden
	width, height := 1280, 800
	if database != nil {
		var savedSize string
		sizeQuery := "SELECT VALUE FROM USER_STATE WHERE \"KEY\" = 'window_size' ORDER BY ID DESC LIMIT 1"
		if cfg.DBEngine == "mysql" {
			sizeQuery = "SELECT VALUE FROM USER_STATE WHERE `KEY` = 'window_size' ORDER BY ID DESC LIMIT 1"
		}
		err = database.SQL.QueryRow(sizeQuery).Scan(&savedSize)
		if err == nil && savedSize != "" {
			fmt.Sscanf(savedSize, "%dx%d", &width, &height)
			log.Printf("Restoring window size: %dx%d", width, height)
		}
	}

	// Maximiert-Status laden
	startState := options.Normal // Default to normal if no DB
	if database != nil {
		startState = options.Maximised
		var savedMax string
		maxQuery := "SELECT VALUE FROM USER_STATE WHERE \"KEY\" = 'window_maximized' ORDER BY ID DESC LIMIT 1"
		if cfg.DBEngine == "mysql" {
			maxQuery = "SELECT VALUE FROM USER_STATE WHERE `KEY` = 'window_maximized' ORDER BY ID DESC LIMIT 1"
		}
		err = database.SQL.QueryRow(maxQuery).Scan(&savedMax)
		if err == nil {
			if savedMax == "false" {
				startState = options.Normal
			}
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
