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
		log.Fatalf("Error: %v", err)
	}
	defer db.Close()

	fmt.Println("=== Checking Silo 9 details ===")
	var silonr int64
	var bez, invAlt, invNeu string
	var qty float64
	err = db.QueryRow("SELECT SILONUMMER, BEZEICHNUNG, INVENTURDATUMALT, INVENTURDATUMNEU, INVENTURFUELLMENGE FROM SILO WHERE ID = 9").
		Scan(&silonr, &bez, &invAlt, &invNeu, &qty)
	if err != nil {
		log.Fatalf("Silo query error: %v", err)
	}
	fmt.Printf("Silo: ID=9, Nr=%d, Bez=%q, Alt=%q, Neu=%q, Qty=%.2f\n", silonr, bez, invAlt, invNeu, qty)

	fmt.Println("\n=== Checking Futter bookings for Silo 9 ===")
	rows, err := db.Query("SELECT ID, LIEFERDATUM, LIEFERMENGE, NETTO, BRUTTO FROM FUTTER WHERE ID_SILO = 9 ORDER BY LIEFERDATUM ASC, ID ASC")
	if err != nil {
		log.Fatalf("Futter query error: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var date string
		var lqty, net, gross float64
		rows.Scan(&id, &date, &lqty, &net, &gross)
		fmt.Printf("  ID: %d | Date: %s | Qty: %.2f | Net: %.2f | Gross: %.2f\n", id, date, lqty, net, gross)
	}
	rows.Close()

	fmt.Println("\n=== Checking BUCHUNG rows for Silo 36 in period ===")
	// Note: using date range from invAlt to invNeu-1
	var sumTB int64
	var count int
	var countNonZero int
	rows, err = db.Query("SELECT ID, BUCHUNGSDATUM, TIERBESTAND, SILONR, FUTTERVERBRAUCHTIER, FUTTERKTAG FROM BUCHUNG WHERE SILONR = ? AND BUCHUNGSDATUM >= ? AND BUCHUNGSDATUM < ? ORDER BY BUCHUNGSDATUM DESC", silonr, invAlt, invNeu)
	if err != nil {
		log.Fatalf("Buchung query error: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		count++
		var id, tb, snr, fvt, fkt int64
		var bdate string
		rows.Scan(&id, &bdate, &tb, &snr, &fvt, &fkt)
		sumTB += tb
		if fvt > 0 || fkt > 0 {
			countNonZero++
		}
		if count <= 10 {
			fmt.Printf("  ID: %5d | Date: %s | TB: %d | Silo: %d | Verbrauch: %d | Ktag: %d\n", id, bdate, tb, snr, fvt, fkt)
		}
	}
	fmt.Printf("Total bookings in range: %d, Non-zero: %d, Sum Tierbestand: %d\n", count, countNonZero, sumTB)
}
