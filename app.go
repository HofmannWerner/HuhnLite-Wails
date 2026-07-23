package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	stdruntime "runtime"

	"huhnlite-wails/backend/api"
	"huhnlite-wails/backend/config"
	"huhnlite-wails/backend/db"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx             context.Context
	database        *db.DB
	ConnectionError string
	apiPort         int
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
	fmt.Printf("[Wails] Entering app.startup. Is database nil? %v\n", a.database == nil)
	file, _ := os.OpenFile("C:\\Users\\hofma\\AppData\\Roaming\\HuhnLite\\debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if file != nil {
		file.WriteString("Entering app.startup\n")
		file.Close()
	}

	// Start the API server synchronously binding the port to prevent frontend race conditions
	if a.database != nil {
		engine := api.StartServer(a.database)
		file, _ = os.OpenFile("C:\\Users\\hofma\\AppData\\Roaming\\HuhnLite\\debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if file != nil {
			file.WriteString(fmt.Sprintf("After StartServer, engine is %v\n", engine != nil))
			file.Close()
		}
		
		port := 8080 // Default port
		if a.database.Config.Port > 0 {
			port = a.database.Config.Port
		}
		
		var listener net.Listener
		var err error
		for attempt := 0; attempt < 20; attempt++ {
			tryPort := port + attempt
			log.Printf("Attempting to start API server on port %d", tryPort)
			listener, err = net.Listen("tcp", fmt.Sprintf(":%d", tryPort))
			
			file, _ = os.OpenFile("C:\\Users\\hofma\\AppData\\Roaming\\HuhnLite\\debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if file != nil {
				file.WriteString(fmt.Sprintf("net.Listen attempt port %d err=%v\n", tryPort, err))
				file.Close()
			}
			
			if err == nil {
				a.database.Config.Port = tryPort
				a.apiPort = tryPort
				log.Printf("Successfully bound API server to port %d", tryPort)
				break
			}
			log.Printf("Port %d in use (%v), trying next port...", tryPort, err)
		}

		if listener != nil {
			go func() {
				file, _ = os.OpenFile("C:\\Users\\hofma\\AppData\\Roaming\\HuhnLite\\debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
				if file != nil {
					file.WriteString(fmt.Sprintf("Starting http.Serve on port %d\n", a.apiPort))
					file.Close()
				}
				if err := http.Serve(listener, engine); err != nil {
					log.Printf("API server stopped: %v", err)
				}
			}()
		} else {
			log.Printf("Failed to bind API server on any port after 20 attempts")
		}
	} else {
		log.Printf("Skipping API server startup: database is nil")
	}

	// Autobackup on start
	if a.database != nil && a.database.Engine != "mysql" {
		ab := a.database.Config.AutoBackup
		if ab == 1 || ab == 3 {
			log.Printf("[Autobackup] Performing startup backup...")
			go func() {
				_, err := api.PerformAutoBackup(a.database, "start")
				if err != nil {
					log.Printf("[Autobackup] Startup backup failed: %v", err)
				} else {
					log.Printf("[Autobackup] Startup backup completed successfully.")
				}
			}()
		}
	}
}

// SaveWindowState saves the current window dimensions to the database
func (a *App) SaveWindowState(ctx context.Context, username string) {
	if a.database == nil {
		return
	}
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

	query := `INSERT INTO USER_STATE (USERNAME, "KEY", VALUE) VALUES (?, 'window_size', ?)
		ON CONFLICT(USERNAME, "KEY") DO UPDATE SET VALUE = excluded.VALUE`
	if a.database.Engine == "mysql" {
		query = `INSERT INTO USER_STATE (USERNAME, ` + "`KEY`" + `, VALUE) VALUES (?, 'window_size', ?)
			ON DUPLICATE KEY UPDATE VALUE = VALUES(VALUE)`
	}

	_, err := a.database.SQL.Exec(query, username, val)
	if err != nil {
		log.Printf("[Wails] Error saving window size: %v", err)
	}

	maxQuery := `INSERT INTO USER_STATE (USERNAME, "KEY", VALUE) VALUES (?, 'window_maximized', ?)
		ON CONFLICT(USERNAME, "KEY") DO UPDATE SET VALUE = excluded.VALUE`
	if a.database.Engine == "mysql" {
		maxQuery = `INSERT INTO USER_STATE (USERNAME, ` + "`KEY`" + `, VALUE) VALUES (?, 'window_maximized', ?)
			ON DUPLICATE KEY UPDATE VALUE = VALUES(VALUE)`
	}

	_, err = a.database.SQL.Exec(maxQuery, username, maxVal)
	if err != nil {
		log.Printf("[Wails] Error saving window maximized: %v", err)
	}
}

// beforeClose is called when the app is about to close
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	log.Printf("[Wails] OnBeforeClose triggered")
	if a.database == nil {
		return false
	}
	// Letzten angemeldeten User finden (oder default)
	var username string
	query := "SELECT USERNAME FROM USER_STATE WHERE \"KEY\" = 'window_size' ORDER BY ID DESC LIMIT 1"
	if a.database.Engine == "mysql" {
		query = "SELECT USERNAME FROM USER_STATE WHERE `KEY` = 'window_size' ORDER BY ID DESC LIMIT 1"
	}
	err := a.database.SQL.QueryRow(query).Scan(&username)
	if err != nil {
		log.Printf("[Wails] Could not find last user for window state: %v", err)
		username = "default"
	}

	log.Printf("[Wails] Saving window state for user: %s", username)
	a.SaveWindowState(ctx, username)

	// Autobackup on exit
	if a.database != nil && a.database.Engine != "mysql" {
		ab := a.database.Config.AutoBackup
		if ab == 2 || ab == 3 {
			log.Printf("[Autobackup] Performing exit backup...")
			_, err := api.PerformAutoBackup(a.database, "exit")
			if err != nil {
				log.Printf("[Autobackup] Exit backup failed: %v", err)
			} else {
				log.Printf("[Autobackup] Exit backup completed successfully.")
			}
		}
	}

	return false // Don't prevent closing
}

// GetDBStatus returns the current database engine and connection info
func (a *App) GetDBStatus() map[string]string {
	if a.database == nil {
		return map[string]string{
			"engine": "offline",
			"host":   "none",
			"error":  a.ConnectionError,
		}
	}
	return map[string]string{
		"engine": a.database.Engine,
		"host":   a.database.ActiveConnStr,
	}
}

// IsTestDB returns whether the database is currently running in test mode
func (a *App) IsTestDB() bool {
	if a.database == nil {
		return false
	}
	return a.database.IsTestMode
}

// GetAPIPort returns the actual port the HTTP API server is bound to
func (a *App) GetAPIPort() int {
	if a.apiPort > 0 {
		return a.apiPort
	}
	if a.database != nil && a.database.Config.Port > 0 {
		return a.database.Config.Port
	}
	return 8080
}

// GetLauncherPort returns the port of the HuhnLite-Select launcher
func (a *App) GetLauncherPort() int {
	if a.database != nil && a.database.Config.LauncherPort > 0 {
		return a.database.Config.LauncherPort
	}
	return 8080
}

// ToggleTestDB toggles between the production and test database
func (a *App) ToggleTestDB(useTest bool) (string, error) {
	if a.database == nil {
		return "", fmt.Errorf("database not initialized")
	}

	var targetConn string
	if useTest {
		targetConn = a.database.Config.DBConnectionTest
		if targetConn == "" {
			return "", fmt.Errorf("no test database connection string defined in settings (db_connection_test)")
		}
	} else {
		targetConn = a.database.Config.DBConnectionString
	}

	err := a.database.SwitchConnection(targetConn, useTest)
	if err != nil {
		return "", err
	}

	// Schema/Migrations auf der neuen Verbindung ausführen
	api.MigrateDB(a.database)

	testVal := 0
	if useTest {
		testVal = 1
	}
	_ = config.SaveTestSetting(testVal, a.database.Engine)

	return a.database.ActiveConnStr, nil
}

// Quit closes the application
func (a *App) Quit() {
	runtime.Quit(a.ctx)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// OpenHelp searches for HuhnLite_[lang].PDF and opens it natively
func (a *App) OpenHelp(lang string) string {
	return a.OpenHelpFile(lang, true)
}

func (a *App) getHelpFilePathsForFile(fileName string) []string {
	var pathsToCheck []string

	// 1. If SQLite, check its directory
	if a.database != nil && a.database.Engine == "sqlite" && a.database.Config.DBConnectionString != "" {
		dbDir := filepath.Dir(a.database.Config.DBConnectionString)
		pathsToCheck = append(pathsToCheck, filepath.Join(dbDir, fileName))
	}

	// 2. Check executable directory (BundleDir)
	if execPath, err := os.Executable(); err == nil {
		bundleDir := filepath.Dir(execPath)
		if filepath.Base(bundleDir) == "MacOS" && filepath.Base(filepath.Dir(bundleDir)) == "Contents" {
			// Check Contents/Resources inside the .app bundle
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

	// 4. Check AppDataDir (Roaming/HuhnLite & Roaming)
	if configDir, err := os.UserConfigDir(); err == nil {
		pathsToCheck = append(pathsToCheck, filepath.Join(configDir, "HuhnLite", fileName))
		pathsToCheck = append(pathsToCheck, filepath.Join(configDir, fileName))
	}

	// 5. Check LOCALAPPDATA environment variable (Local/HuhnLite & Local)
	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		pathsToCheck = append(pathsToCheck, filepath.Join(localAppData, "HuhnLite", fileName))
		pathsToCheck = append(pathsToCheck, filepath.Join(localAppData, fileName))
	}

	// 6. Check APPDATA environment variable (Roaming/HuhnLite & Roaming)
	if appData := os.Getenv("APPDATA"); appData != "" {
		pathsToCheck = append(pathsToCheck, filepath.Join(appData, "HuhnLite", fileName))
		pathsToCheck = append(pathsToCheck, filepath.Join(appData, fileName))
	}

	return pathsToCheck
}

// OpenHelpFile opens the specified help file (HTML or PDF) natively in the OS default handler
func (a *App) OpenHelpFile(lang string, usePDF bool) string {
	if lang == "" {
		lang = "de"
	}
	
	var fileNames []string
	if usePDF {
		fileNames = []string{
			fmt.Sprintf("HuhnLite_%s.PDF", lang),
			fmt.Sprintf("HuhnLite_%s.pdf", lang),
			// fallback to German if different lang requested
			"HuhnLite_de.PDF",
			"HuhnLite_de.pdf",
		}
	} else {
		fileNames = []string{
			fmt.Sprintf("HuhnLite-%s.html", lang),
			"HuhnLite-de.html",
		}
	}

	var targetPath string
	for _, fileName := range fileNames {
		pathsToCheck := a.getHelpFilePathsForFile(fileName)
		for _, p := range pathsToCheck {
			log.Printf("[Help] Checking path for native open: %s", p)
			if _, err := os.Stat(p); err == nil {
				targetPath = p
				break
			}
		}
		if targetPath != "" {
			break
		}
	}

	if targetPath == "" {
		if usePDF {
			return fmt.Sprintf("Die PDF-Hilfedatei konnte nicht gefunden werden.")
		}
		return fmt.Sprintf("Die Hilfedatei konnte nicht gefunden werden.")
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		absPath = targetPath
	}
	
	log.Printf("[Help] Opening help file path natively: %s", absPath)
	
	var cmd *exec.Cmd
	switch stdruntime.GOOS {
	case "windows":
		// On Windows, start can be tricky with quoted arguments. 
		// "start" requires empty title quotes as first argument if target is quoted.
		cmd = exec.Command("cmd", "/c", "start", "", absPath)
	case "darwin":
		cmd = exec.Command("open", absPath)
	default:
		cmd = exec.Command("xdg-open", absPath)
	}
	
	if err := cmd.Start(); err != nil {
		log.Printf("[Help] Error running open command: %v, falling back to Wails browser", err)
		fileURL := "file:///" + filepath.ToSlash(absPath)
		runtime.BrowserOpenURL(a.ctx, fileURL)
	}
	return ""
}

