package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func testConnection(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func main() {
	// Try different ports (3306, 3307) and passwords (turbodiesel, studio)
	passwords := []string{"turbodiesel", "studio"}
	ports := []string{"3306", "3307"}
	var db *sql.DB
	var err error
	var activeDsn string

	fmt.Println("Searching for active MariaDB connection on 192.168.178.92...")
	for _, port := range ports {
		for _, pw := range passwords {
			dsn := fmt.Sprintf("root:%s@tcp(192.168.178.92:%s)/huhnlite?parseTime=true", pw, port)
			fmt.Printf("Trying: %s ... ", dsn)
			db, err = testConnection(dsn)
			if err == nil {
				activeDsn = dsn
				fmt.Println("SUCCESS!")
				break
			} else {
				fmt.Printf("failed: %v\n", err)
			}
		}
		if db != nil {
			break
		}
	}

	if db == nil {
		log.Fatalf("Could not connect to MariaDB on 192.168.178.92 on any tested port/password combination.")
	}
	defer db.Close()

	fmt.Printf("\nUsing active connection: %s\n", activeDsn)

	// 1. Show all tables
	fmt.Println("\n=== List of Tables ===")
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		log.Fatalf("Error showing tables: %v", err)
	}
	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err == nil {
			tables = append(tables, table)
			fmt.Printf("- %s\n", table)
		}
	}
	rows.Close()

	// 2. Check each table
	fmt.Println("\n=== CHECK TABLE status ===")
	for _, table := range tables {
		checkRows, err := db.Query(fmt.Sprintf("CHECK TABLE `%s`", table))
		if err != nil {
			fmt.Printf("Error checking table %s: %v\n", table, err)
			continue
		}
		for checkRows.Next() {
			var tbl, op, msgType, msgText string
			if err := checkRows.Scan(&tbl, &op, &msgType, &msgText); err == nil {
				fmt.Printf("[%s] %s: %s - %s\n", tbl, op, msgType, msgText)
			}
		}
		checkRows.Close()
	}

	// 3. Try to query BUCHUNG/Buchung/buchung specifically
	fmt.Println("\n=== Querying BUCHUNG table ===")
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM buchung").Scan(&count)
	if err != nil {
		fmt.Printf("Failed to SELECT from lowercase buchung: %v\n", err)
	} else {
		fmt.Printf("Successfully selected from lowercase buchung. Count: %d\n", count)
	}
}
