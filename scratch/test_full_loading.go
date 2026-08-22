package main

import (
	"context"
	"fmt"
	"log"

	"huhnlite-wails/backend/config"
	"huhnlite-wails/backend/db"
)

func testLoading(engine string, connStr string, mandant int, isTest int) {
	fmt.Printf("\n==================================================\n")
	fmt.Printf("Testing Engine: %s | Mandant: %d | TestMode: %d\n", engine, mandant, isTest)
	fmt.Printf("Raw Connection String: %s\n", connStr)

	cfg := config.Config{
		DBEngine:           engine,
		DBConnectionString: connStr,
		Mandant:            mandant,
		Test:               isTest,
	}

	cfg.DBConnectionString = config.ApplyMandantToDBConnection(engine, connStr, mandant)
	fmt.Printf("Transformed Connection String: %s\n", cfg.DBConnectionString)

	database, err := db.Connect(cfg)
	if err != nil {
		fmt.Printf("❌ FAILED to connect: %v\n", err)
		return
	}
	defer database.SQL.Close()

	ctx := context.Background()

	// 1. Herden
	herden, err := database.Repo.ListHerden(ctx)
	if err != nil {
		fmt.Printf("❌ ERROR ListHerden: %v\n", err)
	} else {
		fmt.Printf("✅ ListHerden: loaded %d items\n", len(herden))
	}

	// 2. Buchungen
	buchungen, err := database.Repo.ListBuchungen(ctx)
	if err != nil {
		fmt.Printf("❌ ERROR ListBuchungen: %v\n", err)
	} else {
		fmt.Printf("✅ ListBuchungen: loaded %d items\n", len(buchungen))
	}

	// 3. FutterBuchungen
	futter, err := database.Repo.ListFutterBuchungen(ctx)
	if err != nil {
		fmt.Printf("❌ ERROR ListFutterBuchungen: %v\n", err)
	} else {
		fmt.Printf("✅ ListFutterBuchungen: loaded %d items\n", len(futter))
	}

	// 4. Silos
	silos, err := database.Repo.ListSilos(ctx)
	if err != nil {
		fmt.Printf("❌ ERROR ListSilos: %v\n", err)
	} else {
		fmt.Printf("✅ ListSilos: loaded %d items\n", len(silos))
	}

	// 5. Eilager
	eilager, err := database.Repo.ListEilager(ctx)
	if err != nil {
		fmt.Printf("❌ ERROR ListEilager: %v\n", err)
	} else {
		fmt.Printf("✅ ListEilager: loaded %d items\n", len(eilager))
	}
}

func main() {
	log.SetFlags(log.Ltime)

	// MySQL test (prod & test)
	testLoading("mysql", "root:studio@tcp(127.0.0.1:3307)/huhnlite-prod?parseTime=true&allowNativePasswords=true", 1, 0)
	testLoading("mysql", "root:studio@tcp(127.0.0.1:3307)/huhnlite-test?parseTime=true&allowNativePasswords=true", 1, 1)

	// Postgres test (prod & test)
	testLoading("postgres", "postgres://postgres:post@192.168.178.28:5432/huhnlite-prod?sslmode=disable", 1, 0)
	testLoading("postgres", "postgres://postgres:post@192.168.178.28:5432/huhnlite-test?sslmode=disable", 1, 1)
}
