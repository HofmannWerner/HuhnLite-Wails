package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "C:/Users/hofma/GolandProjects/HuhnLite-Wails/HuhnLite.db")
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
