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

	// We check both config files if they exist
	configFiles := []string{"settings.json", "settings_mariadb.json"}

	for _, configFile := range configFiles {
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			continue
		}

		log.Printf("----------------------------------------")
		log.Printf("Processing configuration file: %s", configFile)

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

		// If SQLite and relative path, resolve it relative to CWD
		if cfg.DBEngine == "sqlite" && !filepath.IsAbs(cfg.DBConnectionString) {
			cwd, _ := os.Getwd()
			cfg.DBConnectionString = filepath.Join(cwd, cfg.DBConnectionString)
		}

		log.Printf("Connecting to database: %s (%s)", cfg.DBConnectionString, cfg.DBEngine)
		database, err := db.Connect(cfg)
		if err != nil {
			log.Printf("Failed to connect to database: %v", err)
			continue
		}

		ctx := context.Background()

		// Recreate TRANSLATEFELDNAMEN
		log.Println("Processing TRANSLATEFELDNAMEN...")
		rows, err := database.SQL.QueryContext(ctx, "SELECT ID, FELDNAME FROM FELD_KATALOG")
		if err != nil {
			database.SQL.Close()
			log.Printf("Failed to query FELD_KATALOG: %v", err)
			continue
		}

		recreatedFields := 0
		for rows.Next() {
			var id int64
			var feldname string
			if errScan := rows.Scan(&id, &feldname); errScan == nil {
				upperFeld := strings.ToUpper(feldname)
				trans, exists := fieldTranslations[upperFeld]
				enVal := toCamelCase(feldname)
				itVal := enVal
				if exists {
					enVal = trans[0]
					itVal = trans[1]
				}

				// Insert 'en'
				var count int
				database.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM TRANSLATEFELDNAMEN WHERE ID_FELD_KATALOG = ? AND SPRACHE_KZ = 'en'", id).Scan(&count)
				if count == 0 {
					_, errInsert := database.SQL.ExecContext(ctx, "INSERT INTO TRANSLATEFELDNAMEN (ID_FELD_KATALOG, SPRACHE_KZ, BETREFF, INHALT) VALUES (?, 'en', ?, ?)", id, enVal, enVal)
					if errInsert == nil {
						recreatedFields++
					} else {
						log.Printf("Error inserting en field translation for ID %d: %v", id, errInsert)
					}
				}

				// Insert 'it'
				database.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM TRANSLATEFELDNAMEN WHERE ID_FELD_KATALOG = ? AND SPRACHE_KZ = 'it'", id).Scan(&count)
				if count == 0 {
					_, errInsert := database.SQL.ExecContext(ctx, "INSERT INTO TRANSLATEFELDNAMEN (ID_FELD_KATALOG, SPRACHE_KZ, BETREFF, INHALT) VALUES (?, 'it', ?, ?)", id, itVal, itVal)
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
	// Matches entries like: "ID_HERDEN": {"Herd ID", "ID Gregge"}, even if multiline
	re := regexp.MustCompile(`"([^"]+)"\s*:\s*\{\s*[\s\S]*?"([^"]+)"\s*,\s*[\s\S]*?"([^"]+)"\s*,?\s*\}`)
	matches := re.FindAllStringSubmatch(mapContent, -1)
	for _, m := range matches {
		if len(m) > 3 {
			result[m[1]] = [2]string{m[2], m[3]}
		}
	}
	return result
}

func toCamelCase(s string) string {
	parts := strings.Split(strings.ToLower(s), "_")
	for i, part := range parts {
		if i > 0 {
			parts[i] = strings.Title(part)
		}
	}
	return strings.Join(parts, "")
}
