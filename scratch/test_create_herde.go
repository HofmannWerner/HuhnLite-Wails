package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "C:/Users/hofma/AppData/Roaming/HuhnLite-Wails/HuhnLite_test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Execute INSERT query like CreateHerde
	// HERDENNUMMER: 9999
	// BEZEICHNUNG: "Test Herde"
	// ID_RASSE: 1
	query := `INSERT INTO HERDEN (HERDENNUMMER, BEZEICHNUNG, ID_RASSE, ID_ZUECHTER, ID_EILAGER, ANFANGSBESTAND, EINSTALLDATUM,
                    LEGEDATUM, EINSTALLKOSTEN, ID_SILO, ID_STALL, AKTIV, ALLEBUCHUNGENMITDATUM)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`

	var insertedID int64
	err = db.QueryRow(query, 9999, "Test Herde", 1, 0, 0, 100, "2026-07-09", "2026-07-09", 10.5, 0, 0, 1, 0).Scan(&insertedID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Success! Inserted ID: %d\n", insertedID)
		// Clean up
		db.Exec("DELETE FROM HERDEN WHERE ID = ?", insertedID)
	}
}
