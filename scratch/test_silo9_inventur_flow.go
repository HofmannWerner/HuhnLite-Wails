package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "root:turbodiesel@tcp(192.168.178.60:3307)/huhnlite?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatalf("BeginTx failed: %v", err)
	}
	defer tx.Rollback()

	siloID := int64(9)
	inventurDatum := "2026-04-12"
	inventurMenge := 1000.0 // say, remaining quantity in Brinkmann silo

	fmt.Printf("=== Running Silo 9 Inventory Flow (Dry-Run) ===\n")
	fmt.Printf("Input - SiloID: %d, Date: %s, Qty: %.2f\n", siloID, inventurDatum, inventurMenge)

	// 1. Silo laden
	var silonummer int64
	var bezeichnung string
	var inventurDatumAlt string
	var inventurDatumNeu string
	err = tx.QueryRowContext(ctx, "SELECT SILONUMMER, BEZEICHNUNG, INVENTURDATUMALT, INVENTURDATUMNEU FROM SILO WHERE ID = ?", siloID).Scan(&silonummer, &bezeichnung, &inventurDatumAlt, &inventurDatumNeu)
	if err != nil {
		log.Fatalf("Error loading Silo: %v", err)
	}
	fmt.Printf("Silo Details: Nr=%d, Bez=%q, Alt=%q, Neu=%q\n", silonummer, bezeichnung, inventurDatumAlt, inventurDatumNeu)

	if strings.TrimSpace(inventurDatumAlt) == "" {
		inventurDatumAlt = "0001-01-01"
	}

	// 2. Check 1: Max(Buchungsdatum) der angeschlossenen Herden > Inventurdatum
	var maxBuchungsdatum sql.NullString
	err = tx.QueryRowContext(ctx, "SELECT MAX(BUCHUNGSDATUM) FROM BUCHUNG WHERE SILONR = ?", silonummer).Scan(&maxBuchungsdatum)
	if err != nil {
		log.Fatalf("Error querying max buchung: %v", err)
	}
	maxDateStr := "0001-01-01"
	if maxBuchungsdatum.Valid {
		maxDateStr = maxBuchungsdatum.String
	}
	fmt.Printf("Max booking date for SiloNr %d is %s\n", silonummer, maxDateStr)

	// 3. Letzte Futterlieferung
	var lastF struct {
		ID_SILO         int64
		SILONUMMER      int64
		HERDENR         int64
		ID_PERSON       int64
		LIEFERDATUM     string
		LIEFERMENGE     float64
		PREISDT         float64
		RABATTPROZ      float64
		NETTO           float64
		BRUTTO          float64
		MWSTPROZ        float64
		MWSTKZ          interface{}
		DATUM           string
		ZEITSTEMPEL     string
		ID_FUTTERSORTEN int64
		AW              int64
	}
	err = tx.QueryRowContext(ctx, `
		SELECT ID_SILO, SILONUMMER, HERDENR, ID_PERSON, LIEFERDATUM, LIEFERMENGE, PREISDT, RABATTPROZ, NETTO, BRUTTO, MWSTPROZ, MWSTKZ, DATUM, ZEITSTEMPEL, ID_FUTTERSORTEN, AW
		FROM FUTTER WHERE ID_SILO = ? ORDER BY LIEFERDATUM DESC, ID DESC LIMIT 1`, siloID).
		Scan(&lastF.ID_SILO, &lastF.SILONUMMER, &lastF.HERDENR, &lastF.ID_PERSON, &lastF.LIEFERDATUM, &lastF.LIEFERMENGE, &lastF.PREISDT, &lastF.RABATTPROZ, &lastF.NETTO, &lastF.BRUTTO, &lastF.MWSTPROZ, &lastF.MWSTKZ, &lastF.DATUM, &lastF.ZEITSTEMPEL, &lastF.ID_FUTTERSORTEN, &lastF.AW)
	if err != nil {
		log.Fatalf("Error loading last futter: %v", err)
	}
	fmt.Printf("Last Futter: ID_SILO=%d, LDate=%s, LQty=%.2f, PreisDt=%.2f, Net=%.2f\n",
		lastF.ID_SILO, lastF.LIEFERDATUM, lastF.LIEFERMENGE, lastF.PREISDT, lastF.NETTO)

	// Calculate Netto & Brutto for A
	nettoA := inventurMenge * (lastF.PREISDT / 100.0) * (1.0 - lastF.RABATTPROZ/100.0)
	bruttoA := nettoA * (1.0 + lastF.MWSTPROZ/100.0)

	// mwstKzStr handling
	var mwstKzStr string
	if lastF.MWSTKZ != nil {
		switch v := lastF.MWSTKZ.(type) {
		case []byte:
			mwstKzStr = string(v)
		case string:
			mwstKzStr = v
		default:
			mwstKzStr = fmt.Sprintf("%v", v)
		}
	}

	insertQuery := `
		INSERT INTO FUTTER (ID_SILO, SILONUMMER, HERDENR, ID_PERSON, LIEFERDATUM, LIEFERMENGE, PREISDT, RABATTPROZ, NETTO, BRUTTO, MWSTPROZ, MWSTKZ, DATUM, ZEITSTEMPEL, AW, ID_FUTTERSORTEN)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	resA, err := tx.ExecContext(ctx, insertQuery,
		lastF.ID_SILO, lastF.SILONUMMER, lastF.HERDENR, lastF.ID_PERSON,
		inventurDatum, inventurMenge, lastF.PREISDT, lastF.RABATTPROZ,
		nettoA, bruttoA, lastF.MWSTPROZ, mwstKzStr,
		inventurDatum, inventurDatum+" 12:00:00Z", lastF.AW, lastF.ID_FUTTERSORTEN,
	)
	if err != nil {
		log.Fatalf("Error insert A: %v", err)
	}
	idA, _ := resA.LastInsertId()
	fmt.Printf("Inserted Booking A. ID=%d, Qty=%.2f\n", idA, inventurMenge)

	// Booking B
	tParsed, err := time.Parse("2006-01-02", inventurDatum)
	if err != nil {
		log.Fatalf("Date parse error: %v", err)
	}
	prevDatum := tParsed.AddDate(0, 0, -1).Format("2006-01-02")
	mengeB := -inventurMenge
	nettoB := mengeB * (lastF.PREISDT / 100.0) * (1.0 - lastF.RABATTPROZ/100.0)
	bruttoB := nettoB * (1.0 + lastF.MWSTPROZ/100.0)

	resB, err := tx.ExecContext(ctx, insertQuery,
		lastF.ID_SILO, lastF.SILONUMMER, lastF.HERDENR, lastF.ID_PERSON,
		prevDatum, mengeB, lastF.PREISDT, lastF.RABATTPROZ,
		nettoB, bruttoB, lastF.MWSTPROZ, mwstKzStr,
		prevDatum, prevDatum+" 12:00:00Z", lastF.AW, lastF.ID_FUTTERSORTEN,
	)
	if err != nil {
		log.Fatalf("Error insert B: %v", err)
	}
	idB, _ := resB.LastInsertId()
	fmt.Printf("Inserted Booking B. ID=%d, Qty=%.2f\n", idB, mengeB)

	// Sum deliveries
	var gesamtLiefermenge, gesamtNetto, gesamtBrutto float64
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(LIEFERMENGE), 0.0), COALESCE(SUM(NETTO), 0.0), COALESCE(SUM(BRUTTO), 0.0)
		FROM FUTTER WHERE ID_SILO = ? AND LIEFERDATUM >= ? AND LIEFERDATUM <= ?`,
		siloID, inventurDatumAlt, prevDatum).Scan(&gesamtLiefermenge, &gesamtNetto, &gesamtBrutto)
	if err != nil {
		log.Fatalf("Error summation: %v", err)
	}
	fmt.Printf("Period Summation [%s, %s]: Qty=%.2f, Net=%.2f, Gross=%.2f\n",
		inventurDatumAlt, prevDatum, gesamtLiefermenge, gesamtNetto, gesamtBrutto)

	// Sum total bird days
	var futtertageGesamt int64
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(TIERBESTAND), 0) FROM BUCHUNG
		WHERE SILONR = ? AND BUCHUNGSDATUM >= ? AND BUCHUNGSDATUM <= ?`,
		silonummer, inventurDatumAlt, prevDatum).Scan(&futtertageGesamt)
	if err != nil {
		log.Fatalf("Error query bird days: %v", err)
	}
	fmt.Printf("Bird Days: %d\n", futtertageGesamt)

	if futtertageGesamt <= 0 {
		log.Fatalf("Error: Bird days is <= 0")
	}

	// Compute metrics
	futterverbrauchTierGrams := int64(math.Round((gesamtLiefermenge * 1000.0) / float64(futtertageGesamt)))
	futterkostenTier := gesamtNetto / float64(futtertageGesamt)
	fmt.Printf("Calculated: Consumption/Bird=%d g, Cost/Bird=%.6f EUR\n", futterverbrauchTierGrams, futterkostenTier)

	// Fetch bookings to update
	rows, err := tx.QueryContext(ctx, `
		SELECT ID, TIERBESTAND FROM BUCHUNG
		WHERE SILONR = ? AND BUCHUNGSDATUM >= ? AND BUCHUNGSDATUM <= ?`,
		silonummer, inventurDatumAlt, prevDatum)
	if err != nil {
		log.Fatalf("Error fetching bookings: %v", err)
	}
	defer rows.Close()

	type buchungUpdate struct {
		id          int64
		tierbestand int64
	}
	var updates []buchungUpdate
	for rows.Next() {
		var buID, tb int64
		if err := rows.Scan(&buID, &tb); err == nil {
			updates = append(updates, buchungUpdate{id: buID, tierbestand: tb})
		}
	}
	rows.Close()
	fmt.Printf("Found %d daily bookings in range to update\n", len(updates))

	for _, up := range updates {
		futterKtag := int64(math.Round(futterkostenTier * float64(up.tierbestand)))
		_, err = tx.ExecContext(ctx, `
			UPDATE BUCHUNG
			SET FUTTERVERBRAUCHTIER = ?, FUTTERKTAG = ?
			WHERE ID = ?`,
			futterverbrauchTierGrams, futterKtag, up.id)
		if err != nil {
			log.Fatalf("Failed to update booking ID %d: %v", up.id, err)
		}
	}
	fmt.Printf("Successfully ran the update queries for %d bookings!\n", len(updates))

	fmt.Println("Dry-run completed successfully! Rolling back changes.")
}
