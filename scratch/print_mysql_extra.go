package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "root:studio@tcp(127.0.0.1:3307)/huhnlite?parseTime=true&allowNativePasswords=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening MySQL DB: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT ID, ID_SILO, SILONUMMER, LIEFERDATUM, LIEFERMENGE, NETTO, BRUTTO, ZEITSTEMPEL, MWSTKZ FROM FUTTER WHERE ID >= 1539 OR LIEFERMENGE < 0")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("=== Checking MySQL FUTTER ===")
	count := 0
	for rows.Next() {
		count++
		var id, idSilo, silonr int64
		var lieferDatum, zeitstempel string
		var lieferMenge, netto, brutto float64
		var mwstkz interface{}
		err = rows.Scan(&id, &idSilo, &silonr, &lieferDatum, &lieferMenge, &netto, &brutto, &zeitstempel, &mwstkz)
		if err != nil {
			fmt.Printf("Error scanning MySQL: %v\n", err)
			continue
		}
		fmt.Printf("Futter ID: %d | SiloID: %d | SiloNr: %d | LDate: %s | LQty: %.2f | Net: %.2f | Gross: %.2f | Zeitstempel: %s | MwStKz: %v\n",
			id, idSilo, silonr, lieferDatum, lieferMenge, netto, brutto, zeitstempel, mwstkz)
	}
	fmt.Printf("Found %d rows in MySQL\n", count)
}
