package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "C:/Users/hofma/GolandProjects/HuhnLite-Wails/build/bin/HuhnLite.db")
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	fmt.Println("--- build/bin/HuhnLite.db TRANSLATEFELDNAMEN check ---")
	rowsT, err := db.Query(`
		SELECT fk.ID, fk.FELDNAME, t.SPRACHE_KZ, t.BETREFF, t.INHALT
		FROM FELD_KATALOG fk
		LEFT JOIN TRANSLATEFELDNAMEN t ON fk.ID = t.ID_FELD_KATALOG
		WHERE UPPER(fk.FELDNAME) LIKE '%BUCHUNGSDATUM%' OR UPPER(t.BETREFF) LIKE '%BOOKING%'
	`)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		for rowsT.Next() {
			var id int64
			var fName, lang, betreff, inhalt sql.NullString
			rowsT.Scan(&id, &fName, &lang, &betreff, &inhalt)
			fmt.Printf("FK_ID: %d | Feldname: %s | Lang: %s | Betreff: %s | Inhalt: %s\n",
				id, fName.String, lang.String, betreff.String, inhalt.String)
		}
		rowsT.Close()
	}
}
