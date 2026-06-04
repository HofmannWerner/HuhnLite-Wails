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
		log.Fatalf("Error opening DB: %v", err)
	}
	defer db.Close()

	fmt.Println("=== Columns in SILO ===")
	rows, err := db.Query("PRAGMA table_info(SILO)")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var cid int
		var name, dtype string
		var notnull, pk int
		var dflt_value interface{}
		if err := rows.Scan(&cid, &name, &dtype, &notnull, &dflt_value, &pk); err == nil {
			fmt.Printf("- %s (%s)\n", name, dtype)
		}
	}
	rows.Close()

	fmt.Println("\n=== Sample rows in SILO ===")
	sRows, err := db.Query("SELECT ID, SILONUMMER, BEZEICHNUNG, INVENTURDATUMALT, INVENTURDATUMNEU, INVENTURFUELLMENGE FROM SILO LIMIT 5")
	if err != nil {
		log.Printf("Error querying SILO: %v", err)
	} else {
		defer sRows.Close()
		for sRows.Next() {
			var id int
			var silonummer int
			var bezeichnung string
			var alt, neu string
			var menge float64
			if err := sRows.Scan(&id, &silonummer, &bezeichnung, &alt, &neu, &menge); err == nil {
				fmt.Printf("Silo ID=%d, Nr=%d, Name=%s, Alt=%s, Neu=%s, Menge=%.2f\n", id, silonummer, bezeichnung, alt, neu, menge)
			}
		}
	}

	fmt.Println("\n=== Sample rows in FUTTER ===")
	fRows, err := db.Query("SELECT ID, ID_SILO, SILONUMMER, LIEFERDATUM, LIEFERMENGE, NETTO, BRUTTO, MWSTPROZ, MWSTKZ FROM FUTTER ORDER BY LIEFERDATUM DESC LIMIT 3")
	if err != nil {
		log.Printf("Error querying FUTTER: %v", err)
	} else {
		defer fRows.Close()
		for fRows.Next() {
			var id, idSilo, silonummer int
			var datum string
			var menge, netto, brutto, mwstproz float64
			var mwstkz string
			if err := fRows.Scan(&id, &idSilo, &silonummer, &datum, &menge, &netto, &brutto, &mwstproz, &mwstkz); err == nil {
				fmt.Printf("Futter ID=%d, SiloID=%d, SiloNr=%d, Datum=%s, Menge=%.2f, Netto=%.2f, Brutto=%.2f, MwSt=%.2f (%s)\n", id, idSilo, silonummer, datum, menge, netto, brutto, mwstproz, mwstkz)
			}
		}
	}
}
