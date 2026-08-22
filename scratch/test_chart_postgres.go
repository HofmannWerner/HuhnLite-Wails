package main

import (
	"context"
	"fmt"
	"log"

	"huhnlite-wails/backend/config"
	"huhnlite-wails/backend/db"
	"huhnlite-wails/backend/db/repo"
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

	// 1. GetEggBookingYears
	fmt.Println("=== 1. GetEggBookingYears (id=-1) ===")
	years, err := database.Repo.GetEggBookingYears(ctx, repo.GetEggBookingYearsParams{
		IDHerden:   -1,
		OnlyActive: 0,
	})
	if err != nil {
		fmt.Printf("❌ ERROR GetEggBookingYears: %v\n", err)
	} else {
		fmt.Printf("✅ GetEggBookingYears returned %d years: %v\n", len(years), years)
	}

	// 2. GetEggStatsByHerdeFiltered
	fmt.Println("\n=== 2. GetEggStatsByHerdeFiltered (id=-1, year=2026) ===")
	var yearStr string
	if len(years) > 0 {
		yearStr = fmt.Sprintf("%v", years[0])
	} else {
		yearStr = "2026"
	}

	stats, err := database.Repo.GetEggStatsByHerdeFiltered(ctx, repo.GetEggStatsByHerdeFilteredParams{
		IDHerden:   -1,
		OnlyActive: 0,
		Year:       yearStr,
		Quarter:    0,
		Month:      0,
	})
	if err != nil {
		fmt.Printf("❌ ERROR GetEggStatsByHerdeFiltered: %v\n", err)
	} else {
		fmt.Printf("✅ GetEggStatsByHerdeFiltered returned: %+v\n", stats)
	}

	// 3. GetEggStatsByHerde
	fmt.Println("\n=== 3. GetEggStatsByHerde (id=22) ===")
	eggStats, err := database.Repo.GetEggStatsByHerde(ctx, repo.GetEggStatsByHerdeParams{
		ID:       22,
		IDHerden: 22,
	})
	if err != nil {
		fmt.Printf("❌ ERROR GetEggStatsByHerde: %v\n", err)
	} else {
		fmt.Printf("✅ GetEggStatsByHerde returned: %+v\n", eggStats)
	}

	// 4. GetEggStatsWeeklyByHerde
	fmt.Println("\n=== 4. GetEggStatsWeeklyByHerde (id=22) ===")
	weekly, err := database.Repo.GetEggStatsWeeklyByHerde(ctx, 22)
	if err != nil {
		fmt.Printf("❌ ERROR GetEggStatsWeeklyByHerde: %v\n", err)
	} else {
		fmt.Printf("✅ GetEggStatsWeeklyByHerde returned %d weeks\n", len(weekly))
	}
}
