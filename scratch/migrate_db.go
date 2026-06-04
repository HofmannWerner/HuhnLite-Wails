package main

import (
	"database/sql"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "HuhnLite.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("Starting SQLite database migration...")

	// 1. FUTTER table: add FUTTERTAGE
	_, err = db.Exec("ALTER TABLE FUTTER ADD COLUMN FUTTERTAGE INTEGER NOT NULL DEFAULT 0")
	if err != nil {
		log.Printf("Info/Warning adding FUTTERTAGE: %v", err)
	} else {
		log.Println("✅ Column FUTTERTAGE added to FUTTER table")
	}

	// 2. FUTTER table: drop ZEITSTEMPEL
	_, err = db.Exec("ALTER TABLE FUTTER DROP COLUMN ZEITSTEMPEL")
	if err != nil {
		log.Printf("Info/Warning dropping ZEITSTEMPEL: %v", err)
	} else {
		log.Println("✅ Column ZEITSTEMPEL dropped from FUTTER table")
	}

	log.Println("SQLite database migration completed!")
}
