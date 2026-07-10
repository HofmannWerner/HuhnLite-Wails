package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:studio@tcp(127.0.0.1:3307)/huhnlite?parseTime=true&allowNativePasswords=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Execute INSERT query like CreateHerde in MySQL
	query := `INSERT INTO HERDEN (HERDENNUMMER, BEZEICHNUNG, ID_RASSE, ID_ZUECHTER, ID_EILAGER, ANFANGSBESTAND, EINSTALLDATUM,
                    LEGEDATUM, EINSTALLKOSTEN, ID_SILO, ID_STALL, AKTIV, ALLEBUCHUNGENMITDATUM)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := db.Exec(query, 9998, "Test Herde MySQL", 1, 0, 0, 100, "2026-07-09", "2026-07-09", 10.5, 0, 0, 1, 0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		id, _ := res.LastInsertId()
		fmt.Printf("Success! Inserted ID in MySQL: %d\n", id)
		// Clean up
		db.Exec("DELETE FROM HERDEN WHERE ID = ?", id)
	}
}
