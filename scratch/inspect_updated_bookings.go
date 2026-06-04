package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	dbPath := "C:/Users/hofma/GolandProjects/HuhnLite-Wails/build/bin/HuhnLite.db"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Error opening DB: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT ID, ID_HERDEN, BUCHUNGSDATUM, SILONR, FUTTERVERBRAUCHTIER, FUTTERKTAG, TIERBESTAND 
		FROM BUCHUNG 
		WHERE SILONR = 1 AND (FUTTERVERBRAUCHTIER > 0 OR FUTTERKTAG > 0)
		ORDER BY BUCHUNGSDATUM DESC LIMIT 10`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("=== BUCHUNG rows for Silo 1 in build/bin/HuhnLite.db ===")
	for rows.Next() {
		var id, idHerden, silonr, futterverbrauch, futterktag, tierbestand int64
		var datum string
		if err := rows.Scan(&id, &idHerden, &datum, &silonr, &futterverbrauch, &futterktag, &tierbestand); err == nil {
			fmt.Printf("ID: %d | Herd: %d | Date: %s | Silo: %d | Verbrauch: %d | Ktag: %d | Bestand: %d\n",
				id, idHerden, datum, silonr, futterverbrauch, futterktag, tierbestand)
		} else {
			fmt.Printf("Scan error: %v\n", err)
		}
	}
}
