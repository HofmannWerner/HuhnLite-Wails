package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"huhnlite-wails/backend/config"
	"huhnlite-wails/backend/db"
)

func main() {
	configFiles := []string{"settings.json", "settings_mariadb.json"}

	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			continue
		}

		log.Printf("========================================")
		log.Printf("Checking DB config: %s", configFile)

		file, err := os.Open(configFile)
		if err != nil {
			log.Printf("Failed to open %s: %v", configFile, err)
			continue
		}

		var cfg config.Config
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&cfg); err != nil {
			file.Close()
			log.Printf("Failed to decode config %s: %v", configFile, err)
			continue
		}
		file.Close()

		if cfg.DBEngine == "sqlite" && !filepath.IsAbs(cfg.DBConnectionString) {
			cwd, _ := os.Getwd()
			cfg.DBConnectionString = filepath.Join(cwd, cfg.DBConnectionString)
		}

		log.Printf("Connecting to database: %s", cfg.DBConnectionString)
		database, err := db.Connect(cfg)
		if err != nil {
			log.Printf("Failed to connect to database: %v", err)
			continue
		}

		ctx := context.Background()

		// Count TRANSLATEFELDNAMEN
		log.Println("TRANSLATEFELDNAMEN counts:")
		langs := []string{"de", "en", "it"}
		for _, lang := range langs {
			var count int
			err := database.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM TRANSLATEFELDNAMEN WHERE SPRACHE_KZ = ?", lang).Scan(&count)
			if err != nil {
				log.Printf("  %s: error: %v", lang, err)
			} else {
				log.Printf("  %s: %d records", lang, count)
			}
		}

		// Count FELD_KATALOG
		var fkCount int
		err = database.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM FELD_KATALOG").Scan(&fkCount)
		if err != nil {
			log.Printf("FELD_KATALOG error: %v", err)
		} else {
			log.Printf("FELD_KATALOG: %d records", fkCount)
		}

		// Count UEBERSETZUNGEN
		log.Println("UEBERSETZUNGEN counts:")
		for _, lang := range langs {
			var count int
			err := database.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM UEBERSETZUNGEN WHERE SPRACHE_KZ = ?", lang).Scan(&count)
			if err != nil {
				log.Printf("  %s: error: %v", lang, err)
			} else {
				log.Printf("  %s: %d records", lang, count)
			}
		}

		database.SQL.Close()
	}
}
