package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	dbPath := "C:/Users/hofma/GolandProjects/HuhnLite-Wails/HuhnLite.db"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Error opening DB: %v", err)
	}
	defer db.Close()

	fmt.Println("=== Checking SILO Table ===")
	rowsS, err := db.Query("SELECT ID, SILONUMMER, BEZEICHNUNG, INVENTURDATUMALT, INVENTURDATUMNEU, INVENTURFUELLMENGE FROM SILO")
	if err != nil {
		log.Fatal(err)
	}
	defer rowsS.Close()
	for rowsS.Next() {
		var id, silonr, invMenge int64
		var bez, invAlt, invNeu string
		if err := rowsS.Scan(&id, &silonr, &bez, &invAlt, &invNeu, &invMenge); err == nil {
			fmt.Printf("Silo ID: %d, Nr: %d, Bez: %s, Alt: %s, Neu: %s, Menge: %d\n", id, silonr, bez, invAlt, invNeu, invMenge)
		} else {
			fmt.Printf("Error scanning silo: %v\n", err)
		}
	}

	fmt.Println("\n=== Checking Last 15 Feed Bookings (FUTTER) ===")
	rowsF, err := db.Query(`
		SELECT ID, ID_SILO, SILONUMMER, LIEFERDATUM, LIEFERMENGE, PREISDT, RABATTPROZ, NETTO, BRUTTO, MWSTPROZ, MWSTKZ, DATUM, ZEITSTEMPEL, ID_FUTTERSORTEN, AW 
		FROM FUTTER 
		ORDER BY ID DESC LIMIT 15`)
	if err != nil {
		log.Fatal(err)
	}
	defer rowsF.Close()

	for rowsF.Next() {
		var id, idSilo, silonr, idFutterSorten, aw int64
		var lieferDatum, datum, zeitstempel string
		var lieferMenge, preisdt, rabattproz, netto, brutto, mwstproz float64
		var mwstkz interface{}
		err = rowsF.Scan(&id, &idSilo, &silonr, &lieferDatum, &lieferMenge, &preisdt, &rabattproz, &netto, &brutto, &mwstproz, &mwstkz, &datum, &zeitstempel, &idFutterSorten, &aw)
		if err != nil {
			fmt.Printf("Error scanning futter: %v\n", err)
			continue
		}
		fmt.Printf("Futter ID: %d | SiloID: %d | SiloNr: %d | LDate: %s | LQty: %.2f | Net: %.2f | Gross: %.2f | Sort: %d | MwStKz: %v\n",
			id, idSilo, silonr, lieferDatum, lieferMenge, netto, brutto, idFutterSorten, mwstkz)
	}
}
