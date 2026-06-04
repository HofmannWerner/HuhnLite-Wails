package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/glebarez/go-sqlite"
)

func checkExtra(path string) {
	fmt.Printf("=== Checking DB: %s ===\n", path)
	if _, err := os.Stat(path); err != nil {
		fmt.Printf("File does not exist\n\n")
		return
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT ID, ID_SILO, SILONUMMER, LIEFERDATUM, LIEFERMENGE, PREISDT, RABATTPROZ, NETTO, BRUTTO, MWSTPROZ, MWSTKZ, DATUM, ZEITSTEMPEL, ID_FUTTERSORTEN, AW 
		FROM FUTTER 
		WHERE ID >= 1539`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
		var id, idSilo, silonr, idFutterSorten, aw int64
		var lieferDatum, datum, zeitstempel string
		var lieferMenge, preisdt, rabattproz, netto, brutto, mwstproz float64
		var mwstkz interface{}
		err = rows.Scan(&id, &idSilo, &silonr, &lieferDatum, &lieferMenge, &preisdt, &rabattproz, &netto, &brutto, &mwstproz, &mwstkz, &datum, &zeitstempel, &idFutterSorten, &aw)
		if err != nil {
			fmt.Printf("Error scanning: %v\n", err)
			continue
		}
		fmt.Printf("Futter ID: %d | SiloID: %d | SiloNr: %d | LDate: %s | LQty: %.2f | Net: %.2f | Gross: %.2f | Sort: %d | Zeitstempel: %s | MwStKz: %v\n",
			id, idSilo, silonr, lieferDatum, lieferMenge, netto, brutto, idFutterSorten, zeitstempel, mwstkz)
	}
	fmt.Printf("Found %d rows with ID >= 1539\n\n", count)
}

func main() {
	checkExtra("C:/Users/hofma/GolandProjects/HuhnLite-Wails/HuhnLite.db")
	checkExtra("C:/Users/hofma/GolandProjects/HuhnLite-Wails/build/bin/HuhnLite.db")
	
	configDir, err := os.UserConfigDir()
	if err == nil {
		checkExtra(filepath.Join(configDir, "HuhnLite-Wails", "HuhnLite.db"))
	}
}
