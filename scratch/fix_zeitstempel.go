package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/glebarez/go-sqlite"
)

func migrateDB(path string) {
	fmt.Printf("=== Migrating DB: %s ===\n", path)
	if _, err := os.Stat(path); err != nil {
		fmt.Printf("File does not exist\n\n")
		return
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		fmt.Printf("Error opening DB: %v\n\n", err)
		return
	}
	defer db.Close()

	// Check if ZEITSTEMPEL column exists
	rows, err := db.Query("PRAGMA table_info(FUTTER)")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	hasZeitstempel := false
	for rows.Next() {
		var cid int
		var name, dtype string
		var notnull, pk int
		var dflt_value interface{}
		if err := rows.Scan(&cid, &name, &dtype, &notnull, &dflt_value, &pk); err == nil {
			if name == "ZEITSTEMPEL" {
				hasZeitstempel = true
			}
		}
	}
	rows.Close()

	if !hasZeitstempel {
		fmt.Println("Adding ZEITSTEMPEL column to FUTTER table...")
		_, err = db.Exec("ALTER TABLE FUTTER ADD COLUMN ZEITSTEMPEL VARCHAR(50) DEFAULT '0001-01-01 00:00:00'")
		if err != nil {
			fmt.Printf("Error adding column: %v\n", err)
		} else {
			fmt.Println("✅ Column ZEITSTEMPEL successfully added!")
		}
	} else {
		fmt.Println("ZEITSTEMPEL column already exists.")
	}
	fmt.Println()
}

func main() {
	paths := []string{
		"C:/Users/hofma/GolandProjects/HuhnLite-Wails/HuhnLite.db",
		"C:/Users/hofma/GolandProjects/HuhnLite-Wails/build/bin/HuhnLite.db",
	}

	configDir, err := os.UserConfigDir()
	if err == nil {
		paths = append(paths, filepath.Join(configDir, "HuhnLite-Wails", "HuhnLite.db"))
	}

	for _, path := range paths {
		migrateDB(path)
	}
}
