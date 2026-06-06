package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	paths := []string{
		`C:\Users\hofma\GolandProjects\HuhnLite-Wails\HuhnLite.db`,
		`C:\Users\hofma\AppData\Roaming\HuhnLite-Wails\HuhnLite.db`,
		`C:\Users\hofma\GolandProjects\HuhnLite-Wails\build\bin\HuhnLite.db`,
	}

	for _, p := range paths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			fmt.Printf("File does not exist: %s\n", p)
			continue
		}

		fmt.Printf("\n=== Database: %s ===\n", p)
		dbConn, err := sql.Open("sqlite", p)
		if err != nil {
			log.Printf("Failed to open DB: %v", err)
			continue
		}

		ctx := context.Background()

		// Count TRANSLATEFELDNAMEN
		fmt.Println("TRANSLATEFELDNAMEN counts:")
		langs := []string{"de", "en", "it"}
		for _, lang := range langs {
			var count int
			err := dbConn.QueryRowContext(ctx, "SELECT COUNT(*) FROM TRANSLATEFELDNAMEN WHERE SPRACHE_KZ = ?", lang).Scan(&count)
			if err != nil {
				log.Printf("  %s: error: %v", lang, err)
			} else {
				log.Printf("  %s: %d records", lang, count)
			}
		}

		// Count FELD_KATALOG
		var fkCount int
		err = dbConn.QueryRowContext(ctx, "SELECT COUNT(*) FROM FELD_KATALOG").Scan(&fkCount)
		if err != nil {
			log.Printf("FELD_KATALOG error: %v", err)
		} else {
			log.Printf("FELD_KATALOG: %d records", fkCount)
		}

		// Print max ID_FELD_KATALOG and show if there are duplicates or invalid entries
		var maxID int
		err = dbConn.QueryRowContext(ctx, "SELECT MAX(ID_FELD_KATALOG) FROM TRANSLATEFELDNAMEN").Scan(&maxID)
		if err == nil {
			fmt.Printf("Max ID_FELD_KATALOG: %d\n", maxID)
		}

		dbConn.Close()
	}
}
