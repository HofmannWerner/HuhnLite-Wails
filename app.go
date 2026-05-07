package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"huhnlite-wails/backend/api"
	"huhnlite-wails/backend/db"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx      context.Context
	database *db.DB
}

// NewApp creates a new App application struct
func NewApp(database *db.DB) *App {
	return &App{
		database: database,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Start the API server in a separate goroutine
	go func() {
		engine := api.StartServer(a.database)
		port := 8080 // Default port
		if a.database.Config.Port > 0 {
			port = a.database.Config.Port
		}
		log.Printf("Starting API server on port %d", port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), engine); err != nil {
			log.Printf("Failed to start API server: %v", err)
		}
	}()
}

// SaveWindowState saves the current window dimensions to the database
func (a *App) SaveWindowState(ctx context.Context, username string) {
	if username == "" {
		username = "default"
	}
	w, h := runtime.WindowGetSize(ctx)
	val := fmt.Sprintf("%dx%d", w, h)
	log.Printf("[Wails] Window size to save: %s", val)
	
	isMax := runtime.WindowIsMaximised(ctx)
	maxVal := "false"
	if isMax {
		maxVal = "true"
	}
	log.Printf("[Wails] Window maximized to save: %s", maxVal)

	_, err := a.database.SQL.Exec(`INSERT INTO USER_STATE (USERNAME, KEY, VALUE) VALUES (?, 'window_size', ?)
		ON CONFLICT(USERNAME, KEY) DO UPDATE SET VALUE = excluded.VALUE`,
		username, val)
	if err != nil {
		log.Printf("[Wails] Error saving window size: %v", err)
	}

	_, err = a.database.SQL.Exec(`INSERT INTO USER_STATE (USERNAME, KEY, VALUE) VALUES (?, 'window_maximized', ?)
		ON CONFLICT(USERNAME, KEY) DO UPDATE SET VALUE = excluded.VALUE`,
		username, maxVal)
	if err != nil {
		log.Printf("[Wails] Error saving window maximized: %v", err)
	}
}

// beforeClose is called when the app is about to close
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	log.Printf("[Wails] OnBeforeClose triggered")
	// Letzten angemeldeten User finden (oder default)
	var username string
	err := a.database.SQL.QueryRow("SELECT USERNAME FROM USER_STATE WHERE KEY = 'window_size' ORDER BY ROWID DESC LIMIT 1").Scan(&username)
	if err != nil {
		log.Printf("[Wails] Could not find last user for window state: %v", err)
		username = "default"
	}
	
	log.Printf("[Wails] Saving window state for user: %s", username)
	a.SaveWindowState(ctx, username)
	return false // Don't prevent closing
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
