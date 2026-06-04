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
		log.Fatalf("Error: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT DISTINCT SPRACHE_KZ, COUNT(*) FROM UEBERSETZUNGEN GROUP BY SPRACHE_KZ")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var lang string
			var c int
			rows.Scan(&lang, &c)
			fmt.Printf("UEBERSETZUNGEN - Language: %q, Count: %d\n", lang, c)
		}
	}

	rows2, err := db.Query("SELECT DISTINCT SPRACHE_KZ, COUNT(*) FROM TRANSLATEFELDNAMEN GROUP BY SPRACHE_KZ")
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var lang string
			var c int
			rows2.Scan(&lang, &c)
			fmt.Printf("TRANSLATEFELDNAMEN - Language: %q, Count: %d\n", lang, c)
		}
	}
}
