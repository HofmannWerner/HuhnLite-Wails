package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/glebarez/go-sqlite"
)

func printColumns(path string) {
	fmt.Printf("=== Columns in %s ===\n", path)
	if _, err := os.Stat(path); err != nil {
		fmt.Printf("File does not exist\n\n")
		return
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("PRAGMA table_info(FUTTER)")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, dtype string
		var notnull, pk int
		var dflt_value interface{}
		if err := rows.Scan(&cid, &name, &dtype, &notnull, &dflt_value, &pk); err == nil {
			fmt.Printf("- %s (%s, notnull: %d, default: %v, pk: %d)\n", name, dtype, notnull, dflt_value, pk)
		}
	}
	fmt.Println()
}

func main() {
	printColumns("C:/Users/hofma/GolandProjects/HuhnLite-Wails/HuhnLite.db")
	
	configDir, err := os.UserConfigDir()
	if err == nil {
		printColumns(filepath.Join(configDir, "HuhnLite-Wails", "HuhnLite.db"))
	}
}
