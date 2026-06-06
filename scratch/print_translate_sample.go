package main

import (
	"context"
	"encoding/json"
	"fmt"
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

		fmt.Printf("\n=== Config: %s ===\n", configFile)

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

		database, err := db.Connect(cfg)
		if err != nil {
			log.Printf("Failed to connect to database: %v", err)
			continue
		}

		ctx := context.Background()

		// Print count for each language
		langs := []string{"de", "en", "it"}
		for _, lang := range langs {
			var count int
			database.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM TRANSLATEFELDNAMEN WHERE SPRACHE_KZ = ?", lang).Scan(&count)
			fmt.Printf("Language %s: %d records\n", lang, count)
		}

		// Print some sample records from TRANSLATEFELDNAMEN
		fmt.Println("Sample records:")
		rows, err := database.SQL.QueryContext(ctx, "SELECT ID_FELD_KATALOG, SPRACHE_KZ, BETREFF FROM TRANSLATEFELDNAMEN ORDER BY ID_FELD_KATALOG ASC, SPRACHE_KZ ASC LIMIT 30")
		if err != nil {
			log.Printf("Query failed: %v", err)
			database.SQL.Close()
			continue
		}

		for rows.Next() {
			var id int64
			var lang, betreff string
			rows.Scan(&id, &lang, &betreff)
			fmt.Printf("  ID: %d | Lang: %s | Betreff: %s\n", id, lang, betreff)
		}
		rows.Close()
		database.SQL.Close()
	}
}
