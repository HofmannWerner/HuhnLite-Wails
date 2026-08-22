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
	fmt.Println("\n==================================================")
	fmt.Println("Test 1: ListTierbewegungen (Verluste)")
	fmt.Println("==================================================")
	tb, err := database.Repo.ListTierbewegungen(ctx, "de")
	if err != nil {
		fmt.Printf("❌ ERROR ListTierbewegungen: %v\n", err)
	} else {
		fmt.Printf("✅ ListTierbewegungen returned %d items\n", len(tb))
		for i := 0; i < len(tb) && i < 5; i++ {
			fmt.Printf("  - ID: %d | Typ: %v | Datum: %v | Bew: %v | Grund: %v | Herde: %v\n",
				tb[i].ID, tb[i].Typ, tb[i].Bewegungsdatum, tb[i].Bewegungen, tb[i].GrundText, tb[i].HerdenBezeichnung)
		}
	}

	// Test 2: ListTexte
	fmt.Println("\n==================================================")
	fmt.Println("Test 2: ListTexte & ListTexteByTyp ('V')")
	fmt.Println("==================================================")
	texte, err := database.Repo.ListTexte(ctx, "de")
	if err != nil {
		fmt.Printf("❌ ERROR ListTexte: %v\n", err)
	} else {
		fmt.Printf("✅ ListTexte returned %d items\n", len(texte))
		for i := 0; i < len(texte) && i < 5; i++ {
			fmt.Printf("  - ID: %d | Betreff: %v | SystemKZ: %v\n",
				texte[i].ID, texte[i].Betreff, texte[i].SystemKz)
		}
	}

	// Test 3: ListDynamischeSQL (Reports)
	fmt.Println("\n==================================================")
	fmt.Println("Test 3: Reports / Dynamische SQL Execution")
	fmt.Println("==================================================")
	reports, err := database.Repo.ListDynamischeSQL(ctx)
	if err != nil {
		fmt.Printf("❌ ERROR ListDynamischeSQL: %v\n", err)
	} else {
		fmt.Printf("✅ ListDynamischeSQL returned %d reports\n", len(reports))
		for _, r := range reports {
			fmt.Printf("\n--- Testing Report ID %d: %s ---\n", r.ID, r.Beschreibung)
			
			// Test executing the SQL statement
			fmt.Printf("Executing SQL:\n%s\n", r.Sqlstatement)
			if r.Sqlstatement != "" {
				rows, err := database.SQL.QueryContext(ctx, r.Sqlstatement)
				if err != nil {
					fmt.Printf("❌ SYNTAX ERROR executing Report %d (%s):\n   %v\n", r.ID, r.Beschreibung, err)
				} else {
					cols, _ := rows.Columns()
					fmt.Printf("✅ Report %d executed successfully! Columns: %v\n", r.ID, cols)
					rows.Close()
				}
			} else {
				fmt.Printf("⚠️ Report %d has empty SQL statement!\n", r.ID)
			}
		}
	}
}
