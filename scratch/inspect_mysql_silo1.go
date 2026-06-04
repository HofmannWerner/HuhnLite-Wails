package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func inspectDB(dbPath string) {
	fmt.Printf("\n====== Inspecting DB: %s ======\n", dbPath)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Printf("Error opening DB %s: %v", dbPath, err)
		return
	}
	defer db.Close()

	var id, silonr, invMenge int
	var bez, invAlt, invNeu string
	err = db.QueryRow("SELECT ID, SILONUMMER, BEZEICHNUNG, INVENTURDATUMALT, INVENTURDATUMNEU, INVENTURFUELLMENGE FROM SILO WHERE ID = 1").
		Scan(&id, &silonr, &bez, &invAlt, &invNeu, &invMenge)
	if err != nil {
		log.Printf("Error fetching silo 1: %v", err)
		return
	}
	fmt.Printf("Silo Details -> Nr: %d, Bez: %s, Alt: %s, Neu: %s, Menge: %d\n", silonr, bez, invAlt, invNeu, invMenge)

	// Check test query count (including inactive herds)
	var testCount int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM BUCHUNG B
		JOIN HERDEN H ON B.ID_HERDEN = H.ID
		WHERE H.ID_SILO = ?
		  AND SUBSTR(B.BUCHUNGSDATUM, 1, 10) >= SUBSTR(?, 1, 10)
		  AND SUBSTR(B.BUCHUNGSDATUM, 1, 10) <= ?`,
		1, invAlt, "2026-04-26").Scan(&testCount)
	if err != nil {
		log.Printf("Error test query: %v", err)
		return
	}
	fmt.Printf("Bookings in period (ALL herds): %d\n", testCount)

	// Check if any updated rows exist
	var updatedCount int
	err = db.QueryRow("SELECT COUNT(*) FROM BUCHUNG WHERE SILONR = 1 AND (FUTTERVERBRAUCHTIER > 0 OR FUTTERKTAG > 0)").Scan(&updatedCount)
	if err == nil {
		fmt.Printf("Bookings with SILONR = 1 having FVT or FKT > 0: %d\n", updatedCount)
	}
}

func main() {
	inspectDB("C:/Users/hofma/GolandProjects/HuhnLite-Wails/HuhnLite.db")
	inspectDB("C:/Users/hofma/GolandProjects/HuhnLite-Wails/build/bin/HuhnLite.db")
}
