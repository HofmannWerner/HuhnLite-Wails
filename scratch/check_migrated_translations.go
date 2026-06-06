package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "C:/Users/hofma/AppData/Roaming/HuhnLite-Wails/HuhnLite.db")
	if err != nil {
		log.Fatalf("failed: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT ID_TEXTE, SPRACHE_KZ, BETREFF, INHALT
		FROM UEBERSETZUNGEN
		WHERE UPPER(BETREFF) IN ('VERKÄUFE', 'LAGERWERTE', 'STAMMLISTEN', 'FUTTER-KOSTEN: LIEFERMENGEN & PREISE', 'PRODUKTIONSBERICHT: ZEITRAUM')
		   OR UPPER(INHALT) IN ('VERKÄUFE', 'LAGERWERTE', 'STAMMLISTEN', 'FUTTER-KOSTEN: LIEFERMENGEN & PREISE', 'PRODUKTIONSBERICHT: ZEITRAUM')
		ORDER BY ID_TEXTE ASC, SPRACHE_KZ ASC
	`)
	if err != nil {
		log.Fatalf("query error: %v", err)
	}
	defer rows.Close()

	fmt.Println("Report description translations in UEBERSETZUNGEN:")
	for rows.Next() {
		var id int64
		var lang, betreff, inhalt string
		rows.Scan(&id, &lang, &betreff, &inhalt)
		fmt.Printf("  ID_TEXTE: %d | Lang: %s | Betreff: %s | Inhalt: %s\n",
			id, lang, betreff, inhalt)
	}
}
