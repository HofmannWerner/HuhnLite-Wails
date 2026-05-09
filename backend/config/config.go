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
	Port               int    `json:"port"`          // HTTP Port for server mode or standalone Gin server
}

func LoadConfig() Config {
	appDataDir := ""
	configDir, err := os.UserConfigDir()
	if err == nil {
		// Unter macOS ist das: ~/Library/Application Support/HuhnLite-Wails
		appDataDir = filepath.Join(configDir, "HuhnLite-Wails")
		if err := os.MkdirAll(appDataDir, 0755); err != nil {
			fmt.Printf("ERROR creating AppDataDir: %v\n", err)
		}
		fmt.Printf("AppDataDir: %s\n", appDataDir)

		// Logging in Datei umleiten
		logPath := filepath.Join(appDataDir, "app.log")
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("ERROR opening log file: %v\n", err)
		} else {
			// Sowohl in Konsole als auch in Datei schreiben
			multiWriter := io.MultiWriter(os.Stdout, logFile)
			log.SetOutput(multiWriter)
			log.Printf("--- App gestartet ---")
			log.Printf("Log-Datei: %s", logPath)
		}
	} else {
		fmt.Printf("ERROR getting UserConfigDir: %v\n", err)
	}

	cwd, _ := os.Getwd()
	parent := filepath.Dir(cwd)

	// Verzeichnis der Executable (bzw. bei macOS das Verzeichnis, in dem die .app liegt)
	execPath, err := os.Executable()
	var bundleDir string
	if err == nil {
		bundleDir = filepath.Dir(execPath)
		// Falls wir in einem macOS .app Bundle sind (Contents/MacOS/...)
		if filepath.Base(bundleDir) == "MacOS" && filepath.Base(filepath.Dir(bundleDir)) == "Contents" {
			// Gehe 3 Ebenen hoch: .../HuhnLite-Wails.app -> Verzeichnis, in dem die .app liegt
			bundleDir = filepath.Dir(filepath.Dir(filepath.Dir(bundleDir)))
		}
		fmt.Printf("BundleDir: %s\n", bundleDir)
	}

	// 1. Fallback: Application Support Verzeichnis
	defaultDB := "HuhnLite.db"
	if appDataDir != "" {
		defaultDB = filepath.Join(appDataDir, "HuhnLite.db")
	}

	// 2. Bevorzuge eine bestehende HuhnLite.db neben der App (Portable Mode)
	if bundleDir != "" {
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
		Port:               8080,
	}

	// Name der Einstellungsdatei basierend auf dem Programm-Namen bestimmen
	configName := "settings.json"
	// Wir nutzen den execPath, den wir weiter oben bereits ermittelt haben
	if execPath != "" {
		fullPath := strings.ToLower(execPath)
		// Prüfe ob "mariadb" im Pfad der Executable oder des Bundles vorkommt
		if strings.Contains(fullPath, "mariadb") {
			configName = "settings_mariadb.json"
			log.Printf("MariaDB-Modus erkannt (Dateiname: %s)", configName)
		}
	}

	paths := []string{
		filepath.Join(cwd, configName),
		filepath.Join(parent, configName),
	}
	if bundleDir != "" {
		paths = append(paths, filepath.Join(bundleDir, configName))
	}
	if appDataDir != "" {
		paths = append(paths, filepath.Join(appDataDir, configName))
	}

	for _, p := range paths {
		file, err := os.Open(p)
		if err == nil {
			defer file.Close()
			decoder := json.NewDecoder(file)
			if err := decoder.Decode(&cfg); err == nil {
				log.Printf("Konfiguration aus %s geladen", p)

				// Bei SQLite und relativem Pfad: Pfad relativ zur settings.json auflösen
				if cfg.DBEngine == "sqlite" && !filepath.IsAbs(cfg.DBConnectionString) {
					cfg.DBConnectionString = filepath.Join(filepath.Dir(p), cfg.DBConnectionString)
				}
				return cfg
			}
		}
	}

	log.Printf("Warnung: settings.json nicht gefunden, verwende Defaults (DB: %s)", cfg.DBConnectionString)
	return cfg
}
