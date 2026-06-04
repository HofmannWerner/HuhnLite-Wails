package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "root:turbodiesel@tcp(192.168.178.60:3307)/huhnlite?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}
	defer db.Close()

	var maxID int64
	db.QueryRow("SELECT COALESCE(MAX(ID), 0) FROM FUTTER").Scan(&maxID)
	fmt.Printf("Max FUTTER ID in User's MySQL: %d\n", maxID)

	var maxNegID int64
	db.QueryRow("SELECT COALESCE(MAX(ID), 0) FROM FUTTER WHERE LIEFERMENGE < 0").Scan(&maxNegID)
	fmt.Printf("Max Negative FUTTER ID: %d\n", maxNegID)

	if maxNegID > 0 {
		var id, idSilo, silonr int64
		var ldate string
		var qty, netto, brutto float64
		err := db.QueryRow("SELECT ID, ID_SILO, SILONUMMER, LIEFERDATUM, LIEFERMENGE, NETTO, BRUTTO FROM FUTTER WHERE ID = ?", maxNegID).
			Scan(&id, &idSilo, &silonr, &ldate, &qty, &netto, &brutto)
		if err == nil {
			fmt.Printf("Latest Negative Booking: ID=%d, SiloID=%d, SiloNr=%d, Date=%s, Qty=%.2f, Net=%.2f, Gross=%.2f\n",
				id, idSilo, silonr, ldate, qty, netto, brutto)

			// Find matching positive booking
			var posID int64
			var posDate string
			var posQty float64
			errPos := db.QueryRow("SELECT ID, LIEFERDATUM, LIEFERMENGE FROM FUTTER WHERE ID_SILO = ? AND LIEFERMENGE = ? AND ID != ?", idSilo, -qty, id).Scan(&posID, &posDate, &posQty)
			if errPos == nil {
				fmt.Printf("  Matching Positive: ID=%d, Date=%s, Qty=%.2f\n", posID, posDate, posQty)
			}

			// Silo Alt Date & Neu Date
			var invAlt, invNeu string
			db.QueryRow("SELECT INVENTURDATUMALT, INVENTURDATUMNEU FROM SILO WHERE ID = ?", idSilo).Scan(&invAlt, &invNeu)
			fmt.Printf("  Silo: Alt=%s, Neu=%s\n", invAlt, invNeu)

			// BUCHUNG table status for this Silo
			var updatedCount int
			var zeroCount int
			db.QueryRow("SELECT COUNT(*) FROM BUCHUNG WHERE SILONR = ? AND BUCHUNGSDATUM >= ? AND BUCHUNGSDATUM <= ? AND (FUTTERVERBRAUCHTIER > 0 OR FUTTERKTAG > 0)", silonr, invAlt, ldate).Scan(&updatedCount)
			db.QueryRow("SELECT COUNT(*) FROM BUCHUNG WHERE SILONR = ? AND BUCHUNGSDATUM >= ? AND BUCHUNGSDATUM <= ? AND FUTTERVERBRAUCHTIER = 0 AND FUTTERKTAG = 0", silonr, invAlt, ldate).Scan(&zeroCount)
			fmt.Printf("  BUCHUNG status in [%s, %s] for SiloNr %d: Updated=%d, Zero=%d\n", invAlt, ldate, silonr, updatedCount, zeroCount)
		}
	}
}
