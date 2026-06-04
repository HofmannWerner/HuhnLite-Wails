package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
	db "huhnlite-wails/backend/db/repo"
)

func main() {
	dbPath := "C:/Users/hofma/GolandProjects/HuhnLite-Wails/build/bin/HuhnLite.db"
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Error opening DB: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	queries := db.New(conn)

	res, err := queries.ListFutterBuchungen(ctx)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("=== Checking ListFutterBuchungen Output ===")
	for _, r := range res {
		if r.ID >= 1539 {
			fmt.Printf("ID: %d | IDSilo: %d | Silonummer: %d | Lieferdatum: %s | Liefermenge: %.2f | Netto: %.2f | Brutto: %.2f | SortenID: %d | SortenText: %s\n",
				r.ID, r.IDSilo, r.Silonummer, r.Lieferdatum, r.Liefermenge, r.Netto, r.Brutto, r.IDFuttersorten, r.FuttersorteText.String)
		}
	}
}
