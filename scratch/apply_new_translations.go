package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"
)

type TranslationGroup struct {
	Feldname string
	De       string
	En       string
	It       string
}

func main() {
	translations := []TranslationGroup{
		{"MENGESMALL", "Menge Small", "Quantity Small", "Quantità Small"},
		{"MENGEMEDIUM", "Menge Medium", "Quantity Medium", "Quantità Medium"},
		{"MENGELARGE", "Menge Large", "Quantity Large", "Quantità Large"},
		{"MENGEXL", "Menge XL", "Quantity XL", "Quantità XL"},
		{"PREISSMALL", "Preis Small", "Price Small", "Prezzo Small"},
		{"PREISMEDIUM", "Preis Medium", "Price Medium", "Prezzo Medium"},
		{"PREISLARGE", "Preis Large", "Price Large", "Prezzo Large"},
		{"PREISXL", "Preis XL", "Price XL", "Prezzo XL"},
		{"PASSWORT", "Passwort", "Password", "Password"},
		{"BEWEGUNGSDATUM", "Bewegungsdatum", "Movement Date", "Data Movimento"},
		{"ORT", "Ort", "City", "Città"},
	}

	dbPaths := []string{
		"C:/Users/hofma/AppData/Roaming/HuhnLite-Wails/HuhnLite.db",
		"C:/Users/hofma/GolandProjects/HuhnLite-Wails/HuhnLite.db",
		"C:/Users/hofma/GolandProjects/HuhnLite-Wails/build/bin/HuhnLite.db",
	}

	for _, path := range dbPaths {
		fmt.Printf("=== Updating SQLite: %s ===\n", path)
		db, err := sql.Open("sqlite", path)
		if err == nil {
			updateDB(db, translations)
			db.Close()
		} else {
			log.Printf("Failed to open %s: %v", path, err)
		}
	}

	fmt.Println("=== Updating MySQL ===")
	dbMysql, err := sql.Open("mysql", "root:studio@tcp(127.0.0.1:3307)/huhnlite?parseTime=true")
	if err == nil {
		updateDB(dbMysql, translations)
		dbMysql.Close()
	} else {
		log.Printf("Failed to open MySQL: %v", err)
	}
}

func updateDB(db *sql.DB, translations []TranslationGroup) {
	// Find IDs in FELD_KATALOG
	fkMap := make(map[string]int64)
	rows, err := db.Query("SELECT ID, FELDNAME FROM FELD_KATALOG")
	if err != nil {
		log.Printf("Query error FELD_KATALOG: %v", err)
		return
	}
	for rows.Next() {
		var id int64
		var name string
		rows.Scan(&id, &name)
		fkMap[strings.ToUpper(strings.TrimSpace(name))] = id
	}
	rows.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Printf("Begin transaction error: %v", err)
		return
	}
	defer tx.Rollback()

	inserted := 0
	updated := 0

	for _, t := range translations {
		id, found := fkMap[strings.ToUpper(t.Feldname)]
		if !found {
			log.Printf("Field %s not found in FELD_KATALOG", t.Feldname)
			continue
		}

		langs := []struct {
			code string
			val  string
		}{
			{"de", t.De},
			{"en", t.En},
			{"it", t.It},
		}

		for _, l := range langs {
			var count int
			tx.QueryRow("SELECT COUNT(*) FROM TRANSLATEFELDNAMEN WHERE ID_FELD_KATALOG = ? AND SPRACHE_KZ = ?", id, l.code).Scan(&count)

			if count > 0 {
				_, err = tx.Exec(`
					UPDATE TRANSLATEFELDNAMEN 
					SET BETREFF = ?, INHALT = ? 
					WHERE ID_FELD_KATALOG = ? AND SPRACHE_KZ = ?`,
					l.val, l.val, id, l.code)
				if err == nil {
					updated++
				} else {
					log.Printf("Update error for %s (%s): %v", t.Feldname, l.code, err)
				}
			} else {
				_, err = tx.Exec(`
					INSERT INTO TRANSLATEFELDNAMEN (ID_FELD_KATALOG, SPRACHE_KZ, BETREFF, INHALT) 
					VALUES (?, ?, ?, ?)`,
					id, l.code, l.val, l.val)
				if err == nil {
					inserted++
				} else {
					log.Printf("Insert error for %s (%s): %v", t.Feldname, l.code, err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Commit failed: %v", err)
		return
	}

	fmt.Printf("Done: %d translations inserted, %d updated.\n", inserted, updated)
}
