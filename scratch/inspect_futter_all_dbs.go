package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/glebarez/go-sqlite"
)

func inspectDB(dbPath string) {
	fmt.Printf("\n=========================================\n")
	fmt.Printf("INSPECTING DB: %s\n", dbPath)
	fmt.Printf("=========================================\n")

	if _, err := os.Stat(dbPath); err != nil {
		fmt.Printf("Database file does not exist.\n")
		return
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Printf("Error opening DB: %v\n", err)
		return
	}
	defer db.Close()

	fmt.Println("--- Silo Records ---")
	rowsS, err := db.Query("SELECT ID, SILONUMMER, BEZEICHNUNG, INVENTURDATUMALT, INVENTURDATUMNEU, INVENTURFUELLMENGE FROM SILO")
	if err != nil {
		fmt.Printf("Error querying SILO: %v\n", err)
	} else {
		defer rowsS.Close()
		for rowsS.Next() {
			var id, silonr, invMenge int64
			var bez, invAlt, invNeu string
			if err := rowsS.Scan(&id, &silonr, &bez, &invAlt, &invNeu, &invMenge); err == nil {
				fmt.Printf("  Silo ID: %d | Nr: %d | Bez: %s | Alt: %s | Neu: %s | Menge: %d\n", id, silonr, bez, invAlt, invNeu, invMenge)
			}
		}
	}

	fmt.Println("\n--- Last 10 Feed Bookings (FUTTER) ---")
	rowsF, err := db.Query(`
		SELECT ID, ID_SILO, SILONUMMER, LIEFERDATUM, LIEFERMENGE, NETTO, BRUTTO, ZEITSTEMPEL
		FROM FUTTER 
		ORDER BY ID DESC LIMIT 10`)
	if err != nil {
		fmt.Printf("Error querying FUTTER: %v\n", err)
	} else {
		defer rowsF.Close()
		for rowsF.Next() {
			var id, idSilo, silonr int64
			var lieferDatum, zeitstempel string
			var lieferMenge, netto, brutto float64
			if err := rowsF.Scan(&id, &idSilo, &silonr, &lieferDatum, &lieferMenge, &netto, &brutto, &zeitstempel); err == nil {
				fmt.Printf("  ID: %d | SiloID: %d | SiloNr: %d | LDate: %s | LQty: %.2f | Net: %.2f | Gross: %.2f | Timestamp: %s\n",
					id, idSilo, silonr, lieferDatum, lieferMenge, netto, brutto, zeitstempel)
			}
		}
	}
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
		inspectDB(path)
	}
}
