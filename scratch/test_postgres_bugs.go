package main

import (
	"context"
	"fmt"
	"log"

	"huhnlite-wails/backend/config"
	"huhnlite-wails/backend/db"
)

func main() {
	log.SetFlags(log.Ltime)

	cfg := config.Config{
		DBEngine:           "postgres",
		DBConnectionString: "postgres://postgres:post@192.168.178.28:5432/huhnlite-prod-1?sslmode=disable",
		Mandant:            1,
		Test:               0,
	}

	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("FAILED to connect: %v", err)
	}
	defer database.SQL.Close()

	ctx := context.Background()

	// Test 1: ListTierbewegungen (Verluste)
	fmt.Println("\n=== Test 1: ListTierbewegungen ===")
	tb, err := database.Repo.ListTierbewegungen(ctx, "de")
	if err != nil {
		fmt.Printf("❌ ERROR ListTierbewegungen: %v\n", err)
	} else {
		fmt.Printf("✅ ListTierbewegungen returned %d items\n", len(tb))
		for i := 0; i < len(tb) && i < 3; i++ {
			fmt.Printf("  - [%d] Typ: %s, Datum: %s, Bew: %d, Grund: %s, Herde: %s\n",
				tb[i].ID, tb[i].Typ, tb[i].Bewegungsdatum, tb[i].Bewegungen, tb[i].GrundText, tb[i].HerdenBezeichnung)
		}
	}

	// Test 2: ListTexte
	fmt.Println("\n=== Test 2: ListTexte ===")
	texte, err := database.Repo.ListTexte(ctx, "de")
	if err != nil {
		fmt.Printf("❌ ERROR ListTexte: %v\n", err)
	} else {
		fmt.Printf("✅ ListTexte returned %d items\n", len(texte))
		for i := 0; i < len(texte) && i < 5; i++ {
			fmt.Printf("  - [%d] Betreff: %s, SystemKZ: %d\n",
				texte[i].ID, texte[i].Betreff, texte[i].SystemKz)
		}
	}

	// Test 3: Reports queries / ExecuteReport
	fmt.Println("\n=== Test 3: Reports / Dynamische SQL ===")
	reports, err := database.Repo.ListDynamischeSQL(ctx)
	if err != nil {
		fmt.Printf("❌ ERROR ListDynamischeSQL: %v\n", err)
	} else {
		fmt.Printf("✅ ListDynamischeSQL returned %d reports\n", len(reports))
		for _, r := range reports {
			fmt.Printf("\n--- Testing Report ID %d: %s ---\n", r.ID, r.Titel)
			fmt.Printf("SQL_POSTGRES:\n%s\n", r.SqlPostgres)
			fmt.Printf("SQL_DEFAULT (SQLite/MySQL):\n%s\n", r.Sql)

			queryToRun := r.SqlPostgres
			if queryToRun == "" {
				queryToRun = r.Sql
			}
			if queryToRun != "" {
				rows, err := database.SQL.QueryContext(ctx, queryToRun)
				if err != nil {
					fmt.Printf("❌ ERROR executing report %d: %v\n", r.ID, err)
				} else {
					cols, _ := rows.Columns()
					fmt.Printf("✅ Report %d executed successfully! Columns: %v\n", r.ID, cols)
					rows.Close()
				}
			} else {
				fmt.Printf("⚠️ Report %d has no SQL query!\n", r.ID)
			}
		}
	}
}
