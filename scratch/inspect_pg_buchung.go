package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	dsnPG := "postgres://postgres:post@192.168.178.28:5432/huhnlite-prod-1?sslmode=disable"
	dbPG, err := sql.Open("postgres", dsnPG)
	if err != nil {
		log.Fatalf("PG Open error: %v", err)
	}
	defer dbPG.Close()

	// 1. Count BUCHUNG
	var count int
	err = dbPG.QueryRow("SELECT COUNT(*) FROM BUCHUNG").Scan(&count)
	if err != nil {
		log.Fatalf("Error counting BUCHUNG: %v", err)
	}
	fmt.Printf("BUCHUNG count in Postgres (huhnlite-prod-1): %d\n", count)

	// 2. Sample BUCHUNG rows
	rows, err := dbPG.Query("SELECT id, id_herden, herdennummer, buchungsdatum, vermittelt FROM BUCHUNG LIMIT 10")
	if err != nil {
		log.Fatalf("Error querying BUCHUNG: %v", err)
	}
	defer rows.Close()

	fmt.Println("\nSample BUCHUNG rows:")
	for rows.Next() {
		var id, idHerden int64
		var herdennummer, buchungsdatum, vermittelt sql.NullString
		rows.Scan(&id, &idHerden, &herdennummer, &buchungsdatum, &vermittelt)
		fmt.Printf("  ID: %d | ID_HERDEN: %d | HERDENNUMMER: %s | DATUM: %s | VERMITTELT: '%s'\n",
			id, idHerden, herdennummer.String, buchungsdatum.String, vermittelt.String)
	}

	// 3. Count HERDEN
	var countHerden int
	dbPG.QueryRow("SELECT COUNT(*) FROM HERDEN").Scan(&countHerden)
	fmt.Printf("\nHERDEN count in Postgres: %d\n", countHerden)

	rowsH, err := dbPG.Query("SELECT id, herdennummer, bezeichnung, aktiv FROM HERDEN LIMIT 10")
	if err == nil {
		fmt.Println("Sample HERDEN rows:")
		for rowsH.Next() {
			var id int64
			var nr, bez sql.NullString
			var aktiv sql.NullInt64
			rowsH.Scan(&id, &nr, &bez, &aktiv)
			fmt.Printf("  ID: %d | NR: %s | BEZ: %s | AKTIV: %d\n", id, nr.String, bez.String, aktiv.Int64)
		}
		rowsH.Close()
	}
}
