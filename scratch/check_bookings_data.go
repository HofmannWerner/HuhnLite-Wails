package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func check(path string) {
	fmt.Printf("\n=== checking: %s ===\n", path)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer db.Close()

	var countZero, countNonZero int
	db.QueryRow("SELECT COUNT(*) FROM BUCHUNG WHERE SILONR = 0").Scan(&countZero)
	db.QueryRow("SELECT COUNT(*) FROM BUCHUNG WHERE SILONR > 0").Scan(&countNonZero)

	fmt.Printf("BUCHUNG rows - Silonr=0: %d, Silonr>0: %d\n", countZero, countNonZero)

	// Let's check a few recent ones
	rows, err := db.Query("SELECT ID, ID_HERDEN, BUCHUNGSDATUM, SILONR, FUTTERVERBRAUCHTIER, FUTTERKTAG FROM BUCHUNG ORDER BY ID DESC LIMIT 5")
	if err == nil {
		defer rows.Close()
		fmt.Println("Latest 5 bookings in BUCHUNG:")
		for rows.Next() {
			var id, idHerden, silonr, fvt, fkt int64
			var date string
			rows.Scan(&id, &idHerden, &date, &silonr, &fvt, &fkt)
			fmt.Printf("  ID: %d | Herd: %d | Date: %s | SiloNr: %d | Verbrauch: %d | Ktag: %d\n",
				id, idHerden, date, silonr, fvt, fkt)
		}
	}
}

func main() {
	check("C:/Users/hofma/GolandProjects/HuhnLite-Wails/HuhnLite.db")
}
