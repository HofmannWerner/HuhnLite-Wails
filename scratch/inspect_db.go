package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/glebarez/go-sqlite"
)

func checkDB(path string) {
	fmt.Printf("=== Checking DB: %s ===\n", path)
	if _, err := os.Stat(path); err != nil {
		fmt.Printf("File does not exist: %v\n\n", err)
		return
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		fmt.Printf("Error opening DB: %v\n\n", err)
		return
	}
	defer db.Close()

	fmt.Println("Columns in BUCHUNG:")
	rows, err := db.Query("PRAGMA table_info(BUCHUNG)")
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

	fmt.Println("\nColumns in FIRMENPARAMETER:")
	rows2, err := db.Query("PRAGMA table_info(FIRMENPARAMETER)")
	if err != nil {
		log.Fatal(err)
	}
	defer rows2.Close()

	for rows2.Next() {
		var cid int
		var name, dtype string
		var notnull, pk int
		var dflt_value interface{}
		if err := rows2.Scan(&cid, &name, &dtype, &notnull, &dflt_value, &pk); err == nil {
			fmt.Printf("- %s (%s, notnull: %d, default: %v, pk: %d)\n", name, dtype, notnull, dflt_value, pk)
		}
	}
	fmt.Println()
}

func migrateDB(path string) {
	fmt.Printf("=== Migrating DB: %s ===\n", path)
	if _, err := os.Stat(path); err != nil {
		fmt.Printf("File does not exist: %v\n\n", err)
		return
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		fmt.Printf("Error opening DB: %v\n\n", err)
		return
	}
	defer db.Close()

	// Try adding FUTTERVERBRAUCHTIER to BUCHUNG
	_, err = db.Exec("ALTER TABLE BUCHUNG ADD COLUMN FUTTERVERBRAUCHTIER INTEGER NOT NULL DEFAULT 0")
	if err != nil {
		fmt.Printf("BUCHUNG migration info: %v\n", err)
	} else {
		fmt.Println("Added FUTTERVERBRAUCHTIER to BUCHUNG successfully.")
	}

	// Try adding FUTTERINVENTUR to FIRMENPARAMETER
	_, err = db.Exec("ALTER TABLE FIRMENPARAMETER ADD COLUMN FUTTERINVENTUR INTEGER NOT NULL DEFAULT 0")
	if err != nil {
		fmt.Printf("FIRMENPARAMETER migration info: %v\n", err)
	} else {
		fmt.Println("Added FUTTERINVENTUR to FIRMENPARAMETER successfully.")
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
		checkDB(path)
	}
}
