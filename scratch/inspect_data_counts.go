package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	paths := []string{
		`C:\Users\hofma\GolandProjects\HuhnLite-Wails\HuhnLite.db`,
		`C:\Users\hofma\AppData\Roaming\HuhnLite-Wails\HuhnLite.db`,
		`C:\Users\hofma\GolandProjects\HuhnLite-Wails\build\bin\HuhnLite.db`,
	}

	tables := []string{"HERDEN", "BUCHUNG", "AKTIONEN", "FUTTER", "SILO"}

	// Check SQLite paths
	for _, p := range paths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			continue
		}

		fmt.Printf("\n=== SQLite Database: %s ===\n", p)
		dbConn, err := sql.Open("sqlite", p)
		if err != nil {
			log.Printf("Failed to open DB: %v", err)
			continue
		}

		for _, t := range tables {
			var count int
			err := dbConn.QueryRow("SELECT COUNT(*) FROM " + t).Scan(&count)
			if err != nil {
				fmt.Printf("  %s: error: %v\n", t, err)
			} else {
				fmt.Printf("  %s: %d rows\n", t, count)
			}
		}
		dbConn.Close()
	}

	// Check MySQL Database
	fmt.Printf("\n=== MySQL Database ===\n")
	dbConn, err := sql.Open("mysql", "root:studio@tcp(127.0.0.1:3307)/huhnlite?parseTime=true&allowNativePasswords=true")
	if err != nil {
		log.Printf("Failed to open MySQL DB: %v", err)
		return
	}
	defer dbConn.Close()

	for _, t := range tables {
		var count int
		err := dbConn.QueryRow("SELECT COUNT(*) FROM " + t).Scan(&count)
		if err != nil {
			fmt.Printf("  %s: error: %v\n", t, err)
		} else {
			fmt.Printf("  %s: %d rows\n", t, count)
		}
	}
}
