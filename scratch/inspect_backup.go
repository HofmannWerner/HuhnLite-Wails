package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "C:/Users/hofma/GolandProjects/HuhnLite-Wails/build/bin/backups/HuhnLite2605141024.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Tables in backup database:")
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			fmt.Printf("- %s\n", name)
		}
	}
}
