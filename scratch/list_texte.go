package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	dbPath := "C:/Users/hofma/GolandProjects/HuhnLite-Wails/HuhnLite.db"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT u.ID_TEXTE, u.BETREFF, u.INHALT
		FROM UEBERSETZUNGEN u
		WHERE u.SPRACHE_KZ = 'de'
	`)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var betreff, inhalt string
		if err := rows.Scan(&id, &betreff, &inhalt); err == nil {
			fmt.Printf("ID: %d | Default (de): %s -> %s\n", id, betreff, inhalt)
		}
	}
}
