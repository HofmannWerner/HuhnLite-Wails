package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/glebarez/go-sqlite"
	db "huhnlite-wails/backend/db/repo"
)

func testDB(path string) {
	fmt.Printf("=== Testing DB: %s ===\n", path)
	if _, err := os.Stat(path); err != nil {
		fmt.Printf("File does not exist: %v\n\n", err)
		return
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		fmt.Printf("Error opening DB: %v\n\n", err)
		return
	}
	defer conn.Close()

	queries := db.New(conn)
	res, err := queries.ListFutterBuchungen(context.Background())
	if err != nil {
		fmt.Printf("ERROR running ListFutterBuchungen: %v\n\n", err)
		return
	}

	fmt.Printf("Successfully loaded %d feed bookings.\n", len(res))
	if len(res) > 0 {
		fmt.Println("First few bookings:")
		for i := 0; i < len(res) && i < 3; i++ {
			fmt.Printf("- ID: %d, SiloNr: %d, Date: %s, Qty: %.2f, MwStKz: %v, Sorte: %s\n",
				res[i].ID, res[i].Silonummer, res[i].Lieferdatum, res[i].Liefermenge, res[i].Mwstkz, res[i].FuttersorteText.String)
		}
	}
	fmt.Println()
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
		testDB(path)
	}
}
