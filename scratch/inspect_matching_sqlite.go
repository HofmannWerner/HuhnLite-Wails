package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/glebarez/go-sqlite"
)

func checkDB(dbPath string) {
	fmt.Printf("\nDB: %s\n", dbPath)
	if _, err := os.Stat(dbPath); err != nil {
		fmt.Printf("File not found\n")
		return
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer db.Close()

	// Check if there are any silos
	var siloCount int
	db.QueryRow("SELECT COUNT(*) FROM SILO").Scan(&siloCount)
	fmt.Printf("Silo count: %d\n", siloCount)

	// Check last 2 futter bookings
	fmt.Println("Last 2 FUTTER bookings:")
	rowsF, err := db.Query("SELECT ID, SILONUMMER, LIEFERDATUM, LIEFERMENGE FROM FUTTER ORDER BY ID DESC LIMIT 2")
	if err == nil {
		defer rowsF.Close()
		for rowsF.Next() {
			var id, silonr int64
			var ldate string
			var qty float64
			if err := rowsF.Scan(&id, &silonr, &ldate, &qty); err == nil {
				fmt.Printf("  ID: %d, SiloNr: %d, Date: %s, Qty: %.2f\n", id, silonr, ldate, qty)
			}
		}
	}

	// Check non-zero BUCHUNGs
	var countNonZero int
	db.QueryRow("SELECT COUNT(*) FROM BUCHUNG WHERE FUTTERVERBRAUCHTIER > 0 OR FUTTERKTAG > 0").Scan(&countNonZero)
	fmt.Printf("BUCHUNG rows with consumption/cost > 0: %d\n", countNonZero)

	// Check some BUCHUNGs in general
	var totalBuchung int
	db.QueryRow("SELECT COUNT(*) FROM BUCHUNG").Scan(&totalBuchung)
	fmt.Printf("Total BUCHUNG rows: %d\n", totalBuchung)
}

func main() {
	paths := []string{
		"C:/Users/hofma/GolandProjects/HuhnLite-Wails/HuhnLite.db",
		"C:/Users/hofma/GolandProjects/HuhnLite-Wails/build/bin/HuhnLite.db",
	}

	configDir, err := os.UserConfigDir()
	if err == nil {
		paths = append(paths, filepath.Join(configDir, "HuhnLite-Wails", "HuhnLite.db"))
	}

	for _, path := range paths {
		checkDB(path)
	}
}
