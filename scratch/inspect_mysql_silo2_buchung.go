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

	fmt.Println("=== Checking All Bookings for SiloNr 2 (2026-05-15 to 2026-06-05) ===")
	rows, err := db.Query(`
		SELECT ID, ID_HERDEN, BUCHUNGSDATUM, SILONR, FUTTERVERBRAUCHTIER, FUTTERKTAG, TIERBESTAND
		FROM BUCHUNG 
		WHERE SILONR = 2 AND BUCHUNGSDATUM >= '2026-05-15' AND BUCHUNGSDATUM <= '2026-06-05'
		ORDER BY BUCHUNGSDATUM DESC`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		count++
		var id, idHerden, silonr, futterverbrauch, tb int64
		var futterktag interface{}
		var datum string
		if err := rows.Scan(&id, &idHerden, &datum, &silonr, &futterverbrauch, &futterktag, &tb); err == nil {
			fmt.Printf("  ID: %d | HerdeID: %d | Date: %s | Silonr: %d | Verbrauch: %d | Ktag: %s | Bestand: %d\n",
				id, idHerden, datum, silonr, futterverbrauch, fmt.Sprintf("%s", futterktag), tb)
		}
	}
	fmt.Printf("Total bookings found in range: %d\n", count)
}
