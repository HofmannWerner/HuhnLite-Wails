package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"huhnlite-wails/backend/config"
	"huhnlite-wails/backend/db"
)

func main() {
	log.Println("Loading server.go to extract translations...")
	contentBytes, err := ioutil.ReadFile("backend/api/server.go")
	if err != nil {
		log.Fatalf("Failed to read server.go: %v", err)
	}
	content := string(contentBytes)

	// Extract fieldTranslations map content
	fieldMapContent := extractMapContent(content, "fieldTranslations")
	fieldTranslations := parseTranslationsMap(fieldMapContent)
	log.Printf("Extracted %d field translations", len(fieldTranslations))

	// Extract texteTranslations map content
	textMapContent := extractMapContent(content, "texteTranslations")
	texteTranslations := parseTranslationsMap(textMapContent)
	log.Printf("Extracted %d text translations", len(texteTranslations))

	configFiles := []string{"settings.json", "settings_mariadb.json"}
	
	// Let's also add direct SQLite database files we want to process
	dbPaths := []string{
		`C:\Users\hofma\AppData\Roaming\HuhnLite-Wails\HuhnLite.db`,
	}

	type DBJob struct {
		Name               string
		DBConnectionString string
		DBEngine           string
		Config             *config.Config // if loaded from config
	}

	var jobs []DBJob

	// Add config files as jobs
	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			continue
		}
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

		jobs = append(jobs, DBJob{
			Name:               configFile,
			DBConnectionString: cfg.DBConnectionString,
			DBEngine:           cfg.DBEngine,
			Config:             &cfg,
		})
	}

	// Add direct DB paths
	for _, p := range dbPaths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			continue
		}
		jobs = append(jobs, DBJob{
			Name:               p,
			DBConnectionString: p,
			DBEngine:           "sqlite",
		})
	}

	for _, job := range jobs {
		log.Printf("----------------------------------------")
		log.Printf("Processing database job: %s", job.Name)

		var database *db.DB
		var err error
		if job.Config != nil {
			log.Printf("Connecting to database using config: %s (%s)", job.DBConnectionString, job.DBEngine)
			database, err = db.Connect(*job.Config)
		} else {
			log.Printf("Connecting to database directly: %s (%s)", job.DBConnectionString, job.DBEngine)
			cfg := config.Config{
				DBEngine:           job.DBEngine,
				DBConnectionString: job.DBConnectionString,
			}
			database, err = db.Connect(cfg)
		}
		if err != nil {
			log.Printf("Failed to connect to database: %v", err)
			continue
		}

		ctx := context.Background()

		// Get all FELD_KATALOG names for lookup
		feldKatalog := make(map[int64]string)
		fkRows, err := database.SQL.QueryContext(ctx, "SELECT ID, FELDNAME FROM FELD_KATALOG")
		if err == nil {
			for fkRows.Next() {
				var id int64
				var name string
				if errScan := fkRows.Scan(&id, &name); errScan == nil {
					feldKatalog[id] = name
				}
			}
			fkRows.Close()
		}

		// Recreate TRANSLATEFELDNAMEN based on the 'de' entries
		log.Println("Processing TRANSLATEFELDNAMEN...")
		rows, err := database.SQL.QueryContext(ctx, "SELECT ID_FELD_KATALOG, BETREFF, INHALT FROM TRANSLATEFELDNAMEN WHERE SPRACHE_KZ = 'de'")
		if err != nil {
			database.SQL.Close()
			log.Printf("Failed to query TRANSLATEFELDNAMEN: %v", err)
			continue
		}

		recreatedFields := 0
		for rows.Next() {
			var id int64
			var deBetreff, deInhalt string
			if errScan := rows.Scan(&id, &deBetreff, &deInhalt); errScan == nil {
				enBetreff, itBetreff := deBetreff, deBetreff
				enInhalt, itInhalt := deInhalt, deInhalt

				// 1. Try lookup via FELDNAME
				if name, ok := feldKatalog[id]; ok {
					upperFeld := strings.ToUpper(name)
					if trans, exists := fieldTranslations[upperFeld]; exists {
						enBetreff = trans[0]
						itBetreff = trans[1]
						enInhalt = trans[0]
						itInhalt = trans[1]
					}
				} else {
					// 2. Try lookup via upper-cased deBetreff
					upperFeld := strings.ToUpper(deBetreff)
					if trans, exists := fieldTranslations[upperFeld]; exists {
						enBetreff = trans[0]
						itBetreff = trans[1]
						enInhalt = trans[0]
						itInhalt = trans[1]
					}
				}

				// Insert 'en' if missing
				var count int
				database.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM TRANSLATEFELDNAMEN WHERE ID_FELD_KATALOG = ? AND SPRACHE_KZ = 'en'", id).Scan(&count)
				if count == 0 {
					_, errInsert := database.SQL.ExecContext(ctx, "INSERT INTO TRANSLATEFELDNAMEN (ID_FELD_KATALOG, SPRACHE_KZ, BETREFF, INHALT) VALUES (?, 'en', ?, ?)", id, enBetreff, enInhalt)
					if errInsert == nil {
						recreatedFields++
					} else {
						log.Printf("Error inserting en field translation for ID %d: %v", id, errInsert)
					}
				}

				// Insert 'it' if missing
				database.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM TRANSLATEFELDNAMEN WHERE ID_FELD_KATALOG = ? AND SPRACHE_KZ = 'it'", id).Scan(&count)
				if count == 0 {
					_, errInsert := database.SQL.ExecContext(ctx, "INSERT INTO TRANSLATEFELDNAMEN (ID_FELD_KATALOG, SPRACHE_KZ, BETREFF, INHALT) VALUES (?, 'it', ?, ?)", id, itBetreff, itInhalt)
					if errInsert == nil {
						recreatedFields++
					} else {
						log.Printf("Error inserting it field translation for ID %d: %v", id, errInsert)
					}
				}
			}
		}
		rows.Close()
		log.Printf("Recreated %d records in TRANSLATEFELDNAMEN", recreatedFields)

		// Recreate UEBERSETZUNGEN
		log.Println("Processing UEBERSETZUNGEN...")
		rowsTexte, err := database.SQL.QueryContext(ctx, "SELECT ID_TEXTE, BETREFF, INHALT FROM UEBERSETZUNGEN WHERE SPRACHE_KZ = 'de'")
		if err != nil {
			database.SQL.Close()
			log.Printf("Failed to query UEBERSETZUNGEN: %v", err)
			continue
		}

		recreatedTexts := 0
		for rowsTexte.Next() {
			var idText int64
			var betreff, inhalt string
			if errScan := rowsTexte.Scan(&idText, &betreff, &inhalt); errScan == nil {
				enBetreff := betreff
				itBetreff := betreff
				if trans, found := texteTranslations[betreff]; found {
					enBetreff = trans[0]
					itBetreff = trans[1]
				} else if trans, found := texteTranslations[inhalt]; found {
					enBetreff = trans[0]
					itBetreff = trans[1]
				}

				enInhalt := inhalt
				itInhalt := inhalt
				if trans, found := texteTranslations[inhalt]; found {
					enInhalt = trans[0]
					itInhalt = trans[1]
				} else if trans, found := texteTranslations[betreff]; found {
					enInhalt = trans[0]
					itInhalt = trans[1]
				}

				// Insert 'en'
				var count int
				database.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM UEBERSETZUNGEN WHERE ID_TEXTE = ? AND SPRACHE_KZ = 'en'", idText).Scan(&count)
				if count == 0 {
					_, errInsert := database.SQL.ExecContext(ctx, "INSERT INTO UEBERSETZUNGEN (ID_TEXTE, SPRACHE_KZ, BETREFF, INHALT) VALUES (?, 'en', ?, ?)", idText, enBetreff, enInhalt)
					if errInsert == nil {
						recreatedTexts++
					}
				}

				// Insert 'it'
				database.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM UEBERSETZUNGEN WHERE ID_TEXTE = ? AND SPRACHE_KZ = 'it'", idText).Scan(&count)
				if count == 0 {
					_, errInsert := database.SQL.ExecContext(ctx, "INSERT INTO UEBERSETZUNGEN (ID_TEXTE, SPRACHE_KZ, BETREFF, INHALT) VALUES (?, 'it', ?, ?)", idText, itBetreff, itInhalt)
					if errInsert == nil {
						recreatedTexts++
					}
				}
			}
		}
		rowsTexte.Close()
		log.Printf("Recreated %d records in UEBERSETZUNGEN", recreatedTexts)

		database.SQL.Close()
	}

	log.Printf("----------------------------------------")
	log.Println("Done!")
}

func extractMapContent(source, mapName string) string {
	pattern := mapName + `\s*:=\s*map\[string\]\[2\]string\s*\{([\s\S]*?)\n\t\}`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(source)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func parseTranslationsMap(mapContent string) map[string][2]string {
	result := make(map[string][2]string)
	re := regexp.MustCompile(`"([^"]+)"\s*:\s*\{\s*[\s\S]*?"([^"]+)"\s*,\s*[\s\S]*?"([^"]+)"\s*,?\s*\}`)
	matches := re.FindAllStringSubmatch(mapContent, -1)
	for _, m := range matches {
		if len(m) > 3 {
			result[m[1]] = [2]string{m[2], m[3]}
		}
	}
	return result
}
