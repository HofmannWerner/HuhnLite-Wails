package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
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
//go:embed all:pdfjs
var assets embed.FS

func init() {
	// Ensure .mjs is always served as application/javascript,
	// because Windows registry sometimes lacks this, breaking PDF.js.
	mime.AddExtensionType(".mjs", "application/javascript")
	mime.AddExtensionType(".json", "application/json")
}

func main() {
	// Load config
	cfg := config.LoadConfig()

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
		Title:            "HuhnLite",
		Width:            width,
		Height:           height,
		WindowStartState: startState,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: &HelpAssetHandler{database: database},
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

// HelpAssetHandler serves files from the help/ and pdfjs/ directories natively inside Wails
type HelpAssetHandler struct {
	database *db.DB
}

func (h *HelpAssetHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	log.Printf("[HelpAssetHandler] Request path: %s", path)

	// We only handle requests that start with /help
	if !strings.HasPrefix(path, "/help") {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	// Get resolved help directory
	helpDir := ""
	langs := []string{"de", "en", "it"}
	var pathsToCheck []string

	for _, lang := range langs {
		fileNames := []string{
			fmt.Sprintf("HuhnLite_%s.PDF", lang),
			fmt.Sprintf("HuhnLite_%s.pdf", lang),
		}
		for _, fileName := range fileNames {
			// 1. If SQLite, check its directory
			if h.database != nil && h.database.Engine == "sqlite" && h.database.Config.DBConnectionString != "" {
				dbDir := filepath.Dir(h.database.Config.DBConnectionString)
				pathsToCheck = append(pathsToCheck, filepath.Join(dbDir, fileName))
			}

			// 1.5. Check AppData ConfigDir just in case
			if configDir, err := os.UserConfigDir(); err == nil {
				appDataDir := filepath.Join(configDir, "HuhnLite")
				pathsToCheck = append(pathsToCheck, filepath.Join(appDataDir, fileName))
			}

			// 2. Check executable directory
			if execPath, err := os.Executable(); err == nil {
				bundleDir := filepath.Dir(execPath)
				if filepath.Base(bundleDir) == "MacOS" && filepath.Base(filepath.Dir(bundleDir)) == "Contents" {
					resourcesDir := filepath.Join(filepath.Dir(bundleDir), "Resources")
					pathsToCheck = append(pathsToCheck, filepath.Join(resourcesDir, fileName))
					bundleDir = filepath.Dir(filepath.Dir(filepath.Dir(bundleDir)))
				}
				pathsToCheck = append(pathsToCheck, filepath.Join(bundleDir, fileName))
			}

			// 3. Check CWD
			if cwd, err := os.Getwd(); err == nil {
				pathsToCheck = append(pathsToCheck, filepath.Join(cwd, fileName))
			}
		}
	}

	for _, p := range pathsToCheck {
		if _, err := os.Stat(p); err == nil {
			helpDir = filepath.Dir(p)
			break
		}
	}

	log.Printf("[HelpAssetHandler] Resolved helpDir: %s", helpDir)

	if helpDir == "" {
		log.Printf("[HelpAssetHandler] Help directory not found")
		res.WriteHeader(http.StatusNotFound)
		return
	}

	// Remove /help prefix to get file path relative to helpDir
	relPath := strings.TrimPrefix(path, "/help")
	relPath = strings.TrimPrefix(relPath, "/")
	
	var baseDir string
	var filePath string

	// If requesting pdfjs, prioritize the executable directory/CWD
	if strings.Contains(relPath, "pdfjs") {
		fallbackPdfjs := ""
		if execPath, err := os.Executable(); err == nil {
			execDir := filepath.Dir(execPath)
			if filepath.Base(execDir) == "MacOS" && filepath.Base(filepath.Dir(execDir)) == "Contents" {
				execDir = filepath.Join(filepath.Dir(execDir), "Resources")
			}
			if _, err := os.Stat(filepath.Join(execDir, "pdfjs")); err == nil {
				fallbackPdfjs = execDir
			}
		}
		if fallbackPdfjs == "" {
			if cwd, err := os.Getwd(); err == nil {
				if _, err := os.Stat(filepath.Join(cwd, "pdfjs")); err == nil {
					fallbackPdfjs = cwd
				}
			}
		}
		if fallbackPdfjs != "" {
			baseDir = fallbackPdfjs
			filePath = filepath.Join(fallbackPdfjs, relPath)
			log.Printf("[HelpAssetHandler] PDF.js request, using fallback path: %s (baseDir: %s)", filePath, baseDir)
		}
	}

	if filePath == "" {
		baseDir = helpDir
		filePath = filepath.Join(helpDir, relPath)
		log.Printf("[HelpAssetHandler] Standard request, using path: %s (baseDir: %s)", filePath, baseDir)
	}

	cleanPath := filepath.Clean(filePath)

	// Secure path against traversal attacks (case-insensitive to support Windows drive letter casing)
	cleanPathLower := strings.ToLower(cleanPath)
	baseDirLower := strings.ToLower(baseDir)
	if !strings.HasPrefix(cleanPathLower, baseDirLower) {
		log.Printf("[HelpAssetHandler] Path traversal blocked: %s is not inside baseDir %s", cleanPath, baseDir)
		res.WriteHeader(http.StatusForbidden)
		return
	}

	// Check if file exists on disk
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		// If it's a pdfjs file and doesn't exist on disk, try serving it from embedded assets
		if strings.Contains(relPath, "pdfjs/") || strings.HasPrefix(relPath, "pdfjs") {
			log.Printf("[HelpAssetHandler] File not on disk, attempting to serve from embedded assets: %s", relPath)
			embedPath := filepath.ToSlash(relPath)
			
			// Try to read from embed.FS
			file, err := assets.Open(embedPath)
			if err == nil {
				defer file.Close()
				stat, err := file.Stat()
				if err == nil {
					// Use http.ServeContent which supports Range and MIME types automatically!
					if rs, ok := file.(io.ReadSeeker); ok {
						http.ServeContent(res, req, stat.Name(), stat.ModTime(), rs)
						return
					}
				}
			}
			log.Printf("[HelpAssetHandler] Failed to serve %s from embedded assets", relPath)
		}
	}

	// Serve file using http.ServeFile to ensure correct MIME types and Range header support
	// Provide the clean, resolved path
	http.ServeFile(res, req, cleanPath)
}
