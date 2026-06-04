package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:turbodiesel@tcp(192.168.178.60:3307)/huhnlite?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT ID, BUCHUNGSDATUM, SILONR, TIERBESTAND, FUTTERVERBRAUCHTIER, FUTTERKTAG 
		FROM BUCHUNG 
		WHERE SILONR = 36 
		ORDER BY BUCHUNGSDATUM DESC`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("=== BUCHUNG rows for Silo 36 in MySQL ===")
	count := 0
	for rows.Next() {
		count++
		var id, silonr, tierbestand int64
		var datum string
		var futterverbrauch, futterktag interface{}
		if err := rows.Scan(&id, &datum, &silonr, &tierbestand, &futterverbrauch, &futterktag); err == nil {
			if count <= 15 {
				fmt.Printf("ID: %d | Date: %q | Silo: %d | Bestand: %d | Verbrauch: %v | Ktag: %s\n",
					id, datum, silonr, tierbestand, futterverbrauch, string(futterktag.([]uint8)))
			}
		} else {
			fmt.Printf("Scan error: %v\n", err)
		}
	}
	fmt.Printf("Total Silo 36 bookings: %d\n", count)
}
