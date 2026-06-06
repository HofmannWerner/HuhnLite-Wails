package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "HuhnLite.db")
	if err != nil {
		log.Fatalf("Error opening db: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT ID, BESCHREIBUNG, SQLSTATEMENT, SQLSTATEMENT_NATIVE FROM DYNAMISCHE_SQL")
	if err != nil {
		log.Fatalf("Error querying: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var desc, sqlStmt, sqlNative string
		if err := rows.Scan(&id, &desc, &sqlStmt, &sqlNative); err != nil {
			log.Fatalf("Error scanning: %v", err)
		}
		fmt.Printf("ID: %d | Desc: %s\n", id, desc)
		fmt.Printf("  SQLSTATEMENT:        %q\n", sqlStmt)
		fmt.Printf("  SQLSTATEMENT_NATIVE: %q\n\n", sqlNative)
	}
}
