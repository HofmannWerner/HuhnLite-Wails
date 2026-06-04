package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:studio@tcp(127.0.0.1:3307)/huhnlite?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("=== Checking All Feed Bookings (FUTTER) with ID > 1530 ===")
	rowsF, err := db.Query(`
		SELECT ID, ID_SILO, SILONUMMER, LIEFERDATUM, LIEFERMENGE, PREISDT, RABATTPROZ, NETTO, BRUTTO, DATUM, ZEITSTEMPEL, ID_FUTTERSORTEN, AW 
		FROM FUTTER 
		WHERE ID > 1530
		ORDER BY ID DESC`)
	if err != nil {
		log.Fatal(err)
	}
	defer rowsF.Close()

	for rowsF.Next() {
		var id, idSilo, silonr, idFutterSorten, aw int64
		var lieferDatum, datum, zeitstempel string
		var lieferMenge, preisdt, rabattproz, netto, brutto float64
		err = rowsF.Scan(&id, &idSilo, &silonr, &lieferDatum, &lieferMenge, &preisdt, &rabattproz, &netto, &brutto, &datum, &zeitstempel, &idFutterSorten, &aw)
		if err != nil {
			fmt.Printf("Error scanning futter: %v\n", err)
			continue
		}
		fmt.Printf("Futter ID: %d | SiloID: %d | SiloNr: %d | LDate: %s | LQty: %.2f | Net: %.2f | Gross: %.2f | Sort: %d | Timestamp: %s\n",
			id, idSilo, silonr, lieferDatum, lieferMenge, netto, brutto, idFutterSorten, zeitstempel)
	}
}
