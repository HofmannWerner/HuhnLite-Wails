package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Mode               string `json:"mode"`          // "standalone" or "server"
	DBEngine           string `json:"db_engine"`     // "sqlite" or "mysql"
	DBConnectionString string `json:"db_connection"` // e.g. "HuhnLite.db" or "user:pass@tcp(127.0.0.1:3306)/dbname"
	DBConnectTest      string `json:"db_connect_test"`
	Test               int    `json:"test"`
	Port               int    `json:"port"`          // HTTP Port for server mode or standalone Gin server
	System             int    `json:"system"`        // 1 = Erlaube Bearbeiten von System-Einträgen
}
func LoadConfig() Config {
	cwd, _ := os.Getwd()
	execPath, errExec := os.Executable()
	var bundleDir string
	if errExec == nil {
		bundleDir = filepath.Dir(execPath)
		// Falls wir in einem macOS .app Bundle sind (Contents/MacOS/...)
		if filepath.Base(bundleDir) == "MacOS" && filepath.Base(filepath.Dir(bundleDir)) == "Contents" {
			// Gehe 3 Ebenen hoch: .../HuhnLite-Wails.app -> Verzeichnis, in dem die .app liegt
			bundleDir = filepath.Dir(filepath.Dir(filepath.Dir(bundleDir)))
		}
		fmt.Printf("BundleDir: %s\n", bundleDir)
	}

	appDataDir := ""
	var logFile *os.File
	var logPath string
	configDir, err := os.UserConfigDir()
	if err == nil {
		appDataDir = filepath.Join(configDir, "HuhnLite-Wails")
		if err := os.MkdirAll(appDataDir, 0755); err == nil {
			logPath = filepath.Join(appDataDir, "app.log")
			var errFile error
			logFile, errFile = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if errFile != nil {
				logPath = filepath.Join(appDataDir, fmt.Sprintf("app_%d.log", os.Getpid()))
				logFile, _ = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			}
		} else {
			fmt.Printf("ERROR creating AppDataDir: %v\n", err)
		}
	} else {
		fmt.Printf("ERROR getting UserConfigDir: %v\n", err)
	}

	// Lokal im Verzeichnis der Executable (bzw. CWD als Fallback) loggen
	localLogDir := bundleDir
	if localLogDir == "" {
		localLogDir = cwd
	}
	localLogPath := filepath.Join(localLogDir, "app.log")
	logFileLocal, errLocal := os.OpenFile(localLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if errLocal != nil {
		localLogPath = filepath.Join(localLogDir, fmt.Sprintf("app_%d.log", os.Getpid()))
		logFileLocal, _ = os.OpenFile(localLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	}

	var multiWriter io.Writer
	if logFile != nil && logFileLocal != nil {
		multiWriter = io.MultiWriter(os.Stdout, logFile, logFileLocal)
	} else if logFile != nil {
		multiWriter = io.MultiWriter(os.Stdout, logFile)
	} else if logFileLocal != nil {
		multiWriter = io.MultiWriter(os.Stdout, logFileLocal)
	} else {
		multiWriter = os.Stdout
	}
	log.SetOutput(multiWriter)
	log.Printf("--- App gestartet (PID: %d) ---", os.Getpid())
	if logPath != "" {
		log.Printf("Log-Datei AppData: %s", logPath)
	}
	log.Printf("Log-Datei Lokal: %s", localLogPath)

	parent := filepath.Dir(cwd)

	// Check if running from Program Files or macOS Applications (read-only for standard users)
	isProgramFiles := false
	if bundleDir != "" {
		lowerBundleDir := strings.ToLower(bundleDir)
		if strings.Contains(lowerBundleDir, "program files") || strings.Contains(lowerBundleDir, "programmdateien") || strings.Contains(lowerBundleDir, "/applications") {
			isProgramFiles = true
		}
	}

	if isProgramFiles && appDataDir != "" {
		sourceDirForCopy := bundleDir
		// If running in macOS bundle, copy from Contents/Resources
		if errExec == nil && filepath.Base(filepath.Dir(execPath)) == "MacOS" && filepath.Base(filepath.Dir(filepath.Dir(execPath))) == "Contents" {
			sourceDirForCopy = filepath.Join(filepath.Dir(filepath.Dir(execPath)), "Resources")
		}

		// Files to copy from sourceDirForCopy to appDataDir if they don't exist in appDataDir
		filesToCopy := []string{
			"HuhnLite.db",
			"HuhnLite_test.db",
			"settings.json",
			"settings_mariadb.json",
			"settings_server.json",
			"settings_server_mariadb.json",
		}
		for _, f := range filesToCopy {
			src := filepath.Join(sourceDirForCopy, f)
			dst := filepath.Join(appDataDir, f)
			if _, err := os.Stat(src); err == nil {
				if _, errDst := os.Stat(dst); os.IsNotExist(errDst) {
					log.Printf("Copying %s from bundle to AppData: %s", f, dst)
					if errCopy := copyFile(src, dst); errCopy != nil {
						log.Printf("ERROR copying %s: %v", f, errCopy)
					}
				}
			}
		}
	}

	// 1. Fallback: Application Support Verzeichnis
	defaultDB := "HuhnLite.db"
	if appDataDir != "" {
		defaultDB = filepath.Join(appDataDir, "HuhnLite.db")
	}

	// 2. Bevorzuge eine bestehende HuhnLite.db neben der App (Portable Mode)
	if bundleDir != "" && !isProgramFiles {
		fullPath := filepath.Join(bundleDir, "HuhnLite.db")
		if _, err := os.Stat(fullPath); err == nil {
			fmt.Printf("Found DB at BundleDir: %s\n", fullPath)
			defaultDB = fullPath
		} else {
			fmt.Printf("DB NOT found at BundleDir: %s (%v)\n", fullPath, err)
		}
	}
	// 3. Bevorzuge eine bestehende HuhnLite.db im aktuellen Terminal-Ordner (z.B. für wails dev)
	if cwd != "" && cwd != bundleDir {
		fullPath := filepath.Join(cwd, "HuhnLite.db")
		if _, err := os.Stat(fullPath); err == nil {
			fmt.Printf("Found DB at CWD: %s\n", fullPath)
			defaultDB = fullPath
		} else {
			fmt.Printf("DB NOT found at CWD: %s (%v)\n", fullPath, err)
		}
	}
	fmt.Printf("Selected Default DB: %s\n", defaultDB)

	// Default configuration
	cfg := Config{
		Mode:               "standalone",
		DBEngine:           "sqlite",
		DBConnectionString: defaultDB,
		DBConnectTest:      "",
		Test:               0,
		Port:               8080,
		System:             0,
	}

	// Name der Einstellungsdatei basierend auf dem Programm-Namen bestimmen
	// 1. Priorität: settings.json (Standard/SQLite)
	// 2. Fallback: settings_mariadb.json (falls MariaDB-Modus erkannt)
	configFiles := []string{"settings.json"}

	if execPath != "" {
		fullPath := strings.ToLower(execPath)
		log.Printf("ExecPath: %s", execPath)
		if strings.Contains(fullPath, "server") && strings.Contains(fullPath, "mariadb") {
			// Prioritize Server-MariaDB settings if the binary name suggests it
			configFiles = []string{"settings_server_mariadb.json", "settings.json"}
			log.Printf("Server-MariaDB-Modus erkannt, priorisiere settings_server_mariadb.json")
		} else if strings.Contains(fullPath, "mariadb") {
			// Prioritize MariaDB settings if the binary name suggests it
			configFiles = []string{"settings_mariadb.json", "settings.json"}
			log.Printf("MariaDB-Modus erkannt, priorisiere settings_mariadb.json")
		} else if strings.Contains(fullPath, "server") {
			// Prioritize Server settings if the binary name suggests it
			configFiles = []string{"settings_server.json", "settings.json"}
			log.Printf("Server-Modus erkannt, priorisiere settings_server.json")
		}
	}

	for _, configName := range configFiles {
		var paths []string
		if isProgramFiles && appDataDir != "" {
			// Prioritize AppData settings when running from Program Files (since Program Files is read-only)
			paths = []string{
				filepath.Join(appDataDir, configName),
				filepath.Join(cwd, configName),
				filepath.Join(parent, configName),
			}
		} else {
			paths = []string{
				filepath.Join(cwd, configName),
				filepath.Join(parent, configName),
			}
			if appDataDir != "" {
				paths = append(paths, filepath.Join(appDataDir, configName))
			}
			if bundleDir != "" && !isProgramFiles {
				paths = append(paths, filepath.Join(bundleDir, configName))
			}
		}
		// If running in macOS bundle, check Contents/Resources
		if errExec == nil && filepath.Base(filepath.Dir(execPath)) == "MacOS" && filepath.Base(filepath.Dir(filepath.Dir(execPath))) == "Contents" {
			resourcesDir := filepath.Join(filepath.Dir(filepath.Dir(execPath)), "Resources")
			paths = append(paths, filepath.Join(resourcesDir, configName))
		}

		for _, p := range paths {
			file, err := os.Open(p)
			if err == nil {
				defer file.Close()
				decoder := json.NewDecoder(file)
				if err := decoder.Decode(&cfg); err == nil {
					log.Printf("Konfiguration aus %s geladen (Engine: %s)", p, cfg.DBEngine)
					// Bei SQLite und relativem Pfad: Pfad relativ zur settings.json auflösen
					if cfg.DBEngine == "sqlite" && !filepath.IsAbs(cfg.DBConnectionString) {
						// Clean up Mac-style paths that might have been incorrectly treated as relative on Windows
						if strings.HasPrefix(cfg.DBConnectionString, "/Users/") || strings.HasPrefix(cfg.DBConnectionString, "Users/") {
							cfg.DBConnectionString = filepath.Base(cfg.DBConnectionString)
							log.Printf("Bereinigter Mac-Pfad zu: %s", cfg.DBConnectionString)
						}
						cfg.DBConnectionString = filepath.Join(filepath.Dir(p), cfg.DBConnectionString)
					}
					if cfg.DBEngine == "sqlite" && cfg.DBConnectTest != "" && !filepath.IsAbs(cfg.DBConnectTest) {
						if strings.HasPrefix(cfg.DBConnectTest, "/Users/") || strings.HasPrefix(cfg.DBConnectTest, "Users/") {
							cfg.DBConnectTest = filepath.Base(cfg.DBConnectTest)
						}
						cfg.DBConnectTest = filepath.Join(filepath.Dir(p), cfg.DBConnectTest)
					}
					return cfg
				}
			}
		}
	}
	log.Printf("Warnung: settings.json nicht gefunden, verwende Defaults (DB: %s)", cfg.DBConnectionString)
	return cfg
}

func SaveTestSetting(testVal int, engine string) error {
	configName := "settings.json"
	if engine == "mysql" {
		configName = "settings_mariadb.json"
	}

	cwd, _ := os.Getwd()
	execPath, _ := os.Executable()
	var bundleDir string
	if execPath != "" {
		bundleDir = filepath.Dir(execPath)
		if filepath.Base(bundleDir) == "MacOS" && filepath.Base(filepath.Dir(bundleDir)) == "Contents" {
			bundleDir = filepath.Dir(filepath.Dir(filepath.Dir(bundleDir)))
		}
	}
	parent := filepath.Dir(cwd)
	appDataDir := ""
	configDir, err := os.UserConfigDir()
	if err == nil {
		appDataDir = filepath.Join(configDir, "HuhnLite-Wails")
	}

	paths := []string{
		filepath.Join(cwd, configName),
		filepath.Join(parent, configName),
	}
	if appDataDir != "" {
		paths = append(paths, filepath.Join(appDataDir, configName))
	}
	if bundleDir != "" {
		paths = append(paths, filepath.Join(bundleDir, configName))
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			// Read the file
			data, err := os.ReadFile(p)
			if err != nil {
				continue
			}

			// Unmarshal into a generic map
			var configMap map[string]interface{}
			if err := json.Unmarshal(data, &configMap); err != nil {
				continue
			}

			// Update test field
			configMap["test"] = testVal

			// Marshal back
			newData, err := json.MarshalIndent(configMap, "", "  ")
			if err != nil {
				continue
			}

			// Write back
			err = os.WriteFile(p, newData, 0644)
			if err == nil {
				log.Printf("Successfully saved test=%d to %s", testVal, p)
			}
		}
	}
	return nil
}

// UnmarshalJSON implements a custom JSON unmarshaler to handle multiple types for the "system" field.
func (c *Config) UnmarshalJSON(data []byte) error {
	type Alias Config
	aux := &struct {
		System interface{} `json:"system"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if aux.System != nil {
		switch v := aux.System.(type) {
		case float64:
			c.System = int(v)
		case string:
			if v == "1" || strings.ToLower(v) == "true" {
				c.System = 1
			} else {
				c.System = 0
			}
		case bool:
			if v {
				c.System = 1
			} else {
				c.System = 0
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	return destFile.Sync()
}

