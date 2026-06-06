package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Query SQLite first
	dbSqlite, err := sql.Open("sqlite", "C:/Users/hofma/GolandProjects/HuhnLite-Wails/HuhnLite.db")
	if err == nil {
		defer dbSqlite.Close()
		fmt.Println("=== SQLite Translations ===")
		inspect(dbSqlite)
	} else {
		log.Printf("Failed to open SQLite: %v", err)
	}

	// Query MySQL
	dbMysql, err := sql.Open("mysql", "root:studio@tcp(127.0.0.1:3307)/huhnlite?parseTime=true")
	if err == nil {
		defer dbMysql.Close()
		fmt.Println("\n=== MySQL Translations ===")
		inspect(dbMysql)
	} else {
		log.Printf("Failed to open MySQL: %v", err)
	}
}

func inspect(db *sql.DB) {
	var count int
	_ = db.QueryRow("SELECT COUNT(*) FROM TRANSLATEFELDNAMEN").Scan(&count)
	fmt.Printf("Total rows in TRANSLATEFELDNAMEN: %d\n", count)

	rows, err := db.Query(`
		SELECT fk.FELDNAME, t.SPRACHE_KZ, t.BETREFF, t.INHALT
		FROM TRANSLATEFELDNAMEN t
		JOIN FELD_KATALOG fk ON t.ID_FELD_KATALOG = fk.ID
		WHERE t.SPRACHE_KZ IN ('en', 'it')
		LIMIT 15
	`)
	if err != nil {
		log.Printf("Error querying translations: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name, lang, betreff, inhalt string
		if err := rows.Scan(&name, &lang, &betreff, &inhalt); err == nil {
			fmt.Printf("- Field=%s | Lang=%s | Betreff=%s | Inhalt=%s\n", name, lang, betreff, inhalt)
		}
	}
}
