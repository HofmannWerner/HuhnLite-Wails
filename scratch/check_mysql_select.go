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

	var count int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM BUCHUNG 
		WHERE SILONR = 36 AND BUCHUNGSDATUM >= '1994-06-24 00:00:00' AND BUCHUNGSDATUM <= '2026-03-17'`).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Select count with string dates: %d\n", count)

	// Try query with time.Time or date formatting
	err = db.QueryRow(`
		SELECT COUNT(*) FROM BUCHUNG 
		WHERE SILONR = 36 AND BUCHUNGSDATUM >= DATE('1994-06-24 00:00:00') AND BUCHUNGSDATUM <= DATE('2026-03-17')`).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Select count with DATE(): %d\n", count)
}
