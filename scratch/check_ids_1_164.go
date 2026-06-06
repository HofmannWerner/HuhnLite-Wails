package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "C:/Users/hofma/GolandProjects/HuhnLite-Wails/HuhnLite.db")
	if err != nil {
		log.Fatalf("failed: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT ID_FELD_KATALOG, SPRACHE_KZ, BETREFF, INHALT
		FROM TRANSLATEFELDNAMEN
		WHERE ID_FELD_KATALOG BETWEEN 1 AND 164
		ORDER BY ID_FELD_KATALOG ASC, SPRACHE_KZ ASC
	`)
	if err != nil {
		log.Fatalf("query error: %v", err)
	}
	defer rows.Close()

	fmt.Println("Original translations in HuhnLite.db (1-164):")
	count := 0
	for rows.Next() {
		var fkID int64
		var lang, betreff, inhalt sql.NullString
		rows.Scan(&fkID, &lang, &betreff, &inhalt)
		fmt.Printf("  FK_ID: %d | Lang: %s | Betreff: %s | Inhalt: %s\n",
			fkID, lang.String, betreff.String, inhalt.String)
		count++
	}
	fmt.Printf("Total original rows: %d\n", count)
}
