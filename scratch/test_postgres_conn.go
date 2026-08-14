package main

import (
	"context"
	"fmt"
	"log"

	"huhnlite-wails/backend/config"
	"huhnlite-wails/backend/db"
)

func main() {
	log.Println("=== Testing PostgreSQL Connection & Queries for HuhnLite-Wails ===")

	// Test 1: Connect to huhnlite-prod
	cfgProd := config.Config{
		DBEngine:           "postgres",
		DBConnectionString: "postgres://postgres:post@192.168.178.28:5432/huhnlite-prod?sslmode=disable",
		Test:               0,
	}

	log.Println("1. Connecting to PostgreSQL (huhnlite-prod)...")
	databaseProd, err := db.Connect(cfgProd)
	if err != nil {
		log.Fatalf("FAILED to connect to huhnlite-prod: %v", err)
	}
	defer databaseProd.SQL.Close()

	ctx := context.Background()

	// Query Herden
	herdenProd, err := databaseProd.Repo.ListHerden(ctx)
	if err != nil {
		log.Printf("ERROR listing Herden (prod): %v", err)
	} else {
		log.Printf("SUCCESS: Loaded %d Herden from huhnlite-prod", len(herdenProd))
		for idx, h := range herdenProd {
			if idx < 3 {
				log.Printf("  - Herde ID: %d, Nr: %d, Bezeichnung: %s", h.ID, h.Herdennummer, h.Bezeichnung)
			}
		}
	}

	// Query Company Person
	compProd, err := databaseProd.Repo.GetCompanyPerson(ctx)
	if err != nil {
		log.Printf("INFO GetCompanyPerson (prod): %v", err)
	} else {
		log.Printf("SUCCESS: Company Person (prod): %s %s (%s)", compProd.Name, compProd.Firma, compProd.Ort)
	}

	// Query Silos
	silosProd, err := databaseProd.Repo.ListSilos(ctx)
	if err != nil {
		log.Printf("ERROR listing Silos (prod): %v", err)
	} else {
		log.Printf("SUCCESS: Loaded %d Silos from huhnlite-prod", len(silosProd))
	}

	// Query Eilager
	eilagerProd, err := databaseProd.Repo.ListEilager(ctx)
	if err != nil {
		log.Printf("ERROR listing Eilager (prod): %v", err)
	} else {
		log.Printf("SUCCESS: Loaded %d Eilager from huhnlite-prod", len(eilagerProd))
	}

	fmt.Println("\n--------------------------------------------------")

	// Test 2: Connect to huhnlite-test
	cfgTest := config.Config{
		DBEngine:           "postgres",
		DBConnectionString: "postgres://postgres:post@192.168.178.28:5432/huhnlite-test?sslmode=disable",
		Test:               1,
	}

	log.Println("2. Connecting to PostgreSQL (huhnlite-test)...")
	databaseTest, err := db.Connect(cfgTest)
	if err != nil {
		log.Fatalf("FAILED to connect to huhnlite-test: %v", err)
	}
	defer databaseTest.SQL.Close()

	herdenTest, err := databaseTest.Repo.ListHerden(ctx)
	if err != nil {
		log.Printf("ERROR listing Herden (test): %v", err)
	} else {
		log.Printf("SUCCESS: Loaded %d Herden from huhnlite-test", len(herdenTest))
	}

	log.Println("=== All PostgreSQL integration tests finished successfully! ===")
}
