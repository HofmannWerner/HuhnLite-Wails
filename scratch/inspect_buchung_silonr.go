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

	fmt.Println("=== Checking SILONR in BUCHUNG table ===")
	rows, err := db.Query("SELECT SILONR, COUNT(*) FROM BUCHUNG GROUP BY SILONR")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var silonr, count int64
		if err := rows.Scan(&silonr, &count); err == nil {
			fmt.Printf("SiloNr: %d | Bookings Count: %d\n", silonr, count)
		}
	}
}
