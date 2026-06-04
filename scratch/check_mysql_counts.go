package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:studio@tcp(127.0.0.1:3307)/huhnlite")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tables := []string{"BUCHUNG", "FUTTER", "TIERBEWEGUNGEN", "VERLUSTE"}
	for _, t := range tables {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM " + t).Scan(&count)
		if err != nil {
			fmt.Printf("Error counting %s: %v\n", t, err)
		} else {
			fmt.Printf("Table %s: %d rows\n", t, count)
		}
	}
}
