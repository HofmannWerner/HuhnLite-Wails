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

	var minD, maxD string
	err = db.QueryRow("SELECT MIN(BUCHUNGSDATUM), MAX(BUCHUNGSDATUM) FROM BUCHUNG WHERE SILONR = 36").Scan(&minD, &maxD)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Silo 36 BUCHUNG range in MySQL: [%s, %s]\n", minD, maxD)
}
