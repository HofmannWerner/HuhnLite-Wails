package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "C:/Users/hofma/AppData/Roaming/HuhnLite-Wails/HuhnLite.db")
	if err != nil {
		log.Fatalf("failed: %v", err)
	}
	defer db.Close()

	// Load all FELD_KATALOG records
	fkMap := make(map[string]int64)
	rowsFK, err := db.Query("SELECT ID, FELDNAME FROM FELD_KATALOG")
	if err != nil {
		log.Fatalf("fk error: %v", err)
	}
	for rowsFK.Next() {
		var id int64
		var name string
		rowsFK.Scan(&id, &name)
		fkMap[strings.ToUpper(strings.TrimSpace(name))] = id
	}
	rowsFK.Close()

	// Load German translate field name records between 1 and 164
	rowsT, err := db.Query(`
		SELECT ID_FELD_KATALOG, BETREFF, INHALT 
		FROM TRANSLATEFELDNAMEN 
		WHERE SPRACHE_KZ = 'de' AND ID_FELD_KATALOG BETWEEN 1 AND 164
	`)
	if err != nil {
		log.Fatalf("trans error: %v", err)
	}
	defer rowsT.Close()

	fmt.Println("Mapping orphan TRANSLATEFELDNAMEN (1-164) to active FELD_KATALOG IDs:")
	for rowsT.Next() {
		var oldID int64
		var betreff, inhalt string
		rowsT.Scan(&oldID, &betreff, &inhalt)

		// Try to match German betreff or inhalt with current FELDNAME in FELD_KATALOG
		matchName := ""
		newID := int64(0)

		cleanBetreff := strings.ToUpper(strings.TrimSpace(betreff))
		cleanInhalt := strings.ToUpper(strings.TrimSpace(inhalt))

		if id, found := fkMap[cleanBetreff]; found {
			newID = id
			matchName = cleanBetreff
		} else if id, found := fkMap[cleanInhalt]; found {
			newID = id
			matchName = cleanInhalt
		}

		if newID > 0 {
			fmt.Printf("  Orphan ID_FELD_KATALOG %d (%s) -> Matches FELD_KATALOG ID %d (%s)\n",
				oldID, betreff, newID, matchName)
		} else {
			fmt.Printf("  Orphan ID_FELD_KATALOG %d (%s) -> No match in FELD_KATALOG\n",
				oldID, betreff)
		}
	}
}
