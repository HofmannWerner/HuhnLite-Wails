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

	fmt.Println("=== Checking SILONR, FUTTERVERBRAUCHTIER, FUTTERKTAG in MySQL BUCHUNG ===")
	
	var countZero, countNonNull, countTotal int
	err = db.QueryRow("SELECT COUNT(*) FROM BUCHUNG").Scan(&countTotal)
	if err != nil {
		log.Fatal(err)
	}
	
	err = db.QueryRow("SELECT COUNT(*) FROM BUCHUNG WHERE FUTTERVERBRAUCHTIER = 0 AND FUTTERKTAG = 0").Scan(&countZero)
	if err != nil {
		log.Fatal(err)
	}
	
	err = db.QueryRow("SELECT COUNT(*) FROM BUCHUNG WHERE FUTTERVERBRAUCHTIER > 0 OR FUTTERKTAG > 0").Scan(&countNonNull)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Total MySQL bookings: %d\n", countTotal)
	fmt.Printf("Bookings with 0 values: %d\n", countZero)
	fmt.Printf("Bookings with non-zero values: %d\n", countNonNull)
	
	if countNonNull > 0 {
		fmt.Println("\nSample non-zero bookings in MySQL:")
		rows, err := db.Query("SELECT ID, ID_HERDEN, BUCHUNGSDATUM, SILONR, FUTTERVERBRAUCHTIER, FUTTERKTAG FROM BUCHUNG WHERE FUTTERVERBRAUCHTIER > 0 OR FUTTERKTAG > 0 LIMIT 15")
		if err != nil {
			log.Fatalf("Query error: %v", err)
		}
		defer rows.Close()
		for rows.Next() {
			var id, idHerden, silonr, futterverbrauch, futterktag int64
			var datum string
			if err := rows.Scan(&id, &idHerden, &datum, &silonr, &futterverbrauch, &futterktag); err == nil {
				fmt.Printf("  ID: %d | HerdeID: %d | Date: %s | Silonr: %d | Verbrauch: %d | Ktag: %d\n",
					id, idHerden, datum, silonr, futterverbrauch, futterktag)
			} else {
				fmt.Printf("  Scan error: %v\n", err)
			}
		}
	}
}
