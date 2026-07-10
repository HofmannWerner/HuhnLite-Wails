package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "C:/Users/hofma/AppData/Roaming/HuhnLite-Wails/HuhnLite_test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// List herds
	rows, err := db.Query("SELECT ID, HERDENNUMMER, BEZEICHNUNG, ID_RASSE, AKTIV FROM HERDEN")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Herds in Test Database:")
	for rows.Next() {
		var id, herdennummer, idRasse, aktiv int64
		var bezeichnung string
		if err := rows.Scan(&id, &herdennummer, &bezeichnung, &idRasse, &aktiv); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ID: %d, Nr: %d, Bezeichnung: %s, Rasse: %d, Aktiv: %d\n", id, herdennummer, bezeichnung, idRasse, aktiv)
	}
}
