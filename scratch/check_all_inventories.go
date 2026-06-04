package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"
)

func checkSQLite(path string) {
	fmt.Printf("\n=== SQLite: %s ===\n", path)
	if _, err := os.Stat(path); err != nil {
		fmt.Println("Not found")
		return
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer db.Close()

	checkDB(db)
}

func checkMySQL() {
	fmt.Println("\n=== MySQL: root:studio@tcp(127.0.0.1:3307)/huhnlite ===")
	db, err := sql.Open("mysql", "root:studio@tcp(127.0.0.1:3307)/huhnlite?parseTime=true")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer db.Close()

	checkDB(db)
}

func checkDB(db *sql.DB) {
	// Find negative deliveries (Booking B)
	rows, err := db.Query("SELECT ID, ID_SILO, SILONUMMER, LIEFERDATUM, LIEFERMENGE, NETTO, BRUTTO, ZEITSTEMPEL FROM FUTTER WHERE LIEFERMENGE < 0")
	if err != nil {
		fmt.Printf("Error querying negative FUTTER: %v\n", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
		var id, idSilo, silonr int64
		var ldate, zeitstempel string
		var qty, netto, brutto float64
		if err := rows.Scan(&id, &idSilo, &silonr, &ldate, &qty, &netto, &brutto, &zeitstempel); err == nil {
			fmt.Printf("Negative Booking: ID=%d, SiloID=%d, SiloNr=%d, Date=%s, Qty=%.2f, Net=%.2f, Gross=%.2f, Time=%s\n",
				id, idSilo, silonr, ldate, qty, netto, brutto, zeitstempel)

			// Find the matching positive booking
			var posID int64
			var posDate string
			var posQty float64
			errPos := db.QueryRow("SELECT ID, LIEFERDATUM, LIEFERMENGE FROM FUTTER WHERE ID_SILO = ? AND LIEFERMENGE = ? AND ID != ?", idSilo, -qty, id).Scan(&posID, &posDate, &posQty)
			if errPos == nil {
				fmt.Printf("  Matching Positive: ID=%d, Date=%s, Qty=%.2f\n", posID, posDate, posQty)
			} else {
				fmt.Printf("  No matching positive booking found: %v\n", errPos)
			}

			// Let's check Silo info
			var silonrS, invMenge int64
			var bez, invAlt, invNeu string
			errS := db.QueryRow("SELECT SILONUMMER, BEZEICHNUNG, INVENTURDATUMALT, INVENTURDATUMNEU, INVENTURFUELLMENGE FROM SILO WHERE ID = ?", idSilo).Scan(&silonrS, &bez, &invAlt, &invNeu, &invMenge)
			if errS == nil {
				fmt.Printf("  Silo: Nr=%d, Bez=%s, Alt=%s, Neu=%s, Qty=%d\n", silonrS, bez, invAlt, invNeu, invMenge)
			}

			// Let's check BUCHUNG table updates for this SiloNr in the range [invAlt, Date]
			var updatedCount int
			var zeroCount int
			errB := db.QueryRow("SELECT COUNT(*) FROM BUCHUNG WHERE SILONR = ? AND BUCHUNGSDATUM >= ? AND BUCHUNGSDATUM <= ? AND (FUTTERVERBRAUCHTIER > 0 OR FUTTERKTAG > 0)", silonr, invAlt, ldate).Scan(&updatedCount)
			errB2 := db.QueryRow("SELECT COUNT(*) FROM BUCHUNG WHERE SILONR = ? AND BUCHUNGSDATUM >= ? AND BUCHUNGSDATUM <= ? AND FUTTERVERBRAUCHTIER = 0 AND FUTTERKTAG = 0", silonr, invAlt, ldate).Scan(&zeroCount)
			if errB == nil && errB2 == nil {
				fmt.Printf("  BUCHUNG in range [%s, %s] for SiloNr %d: Updated=%d, Remaining Zero=%d (Total=%d)\n",
					invAlt, ldate, silonr, updatedCount, zeroCount, updatedCount+zeroCount)
			} else {
				fmt.Printf("  Error checking BUCHUNG updates: %v, %v\n", errB, errB2)
			}
		}
	}
	if count == 0 {
		fmt.Println("No negative feed bookings found.")
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
		checkSQLite(path)
	}

	checkMySQL()
}
