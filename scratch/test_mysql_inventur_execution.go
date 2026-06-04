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
	dsn := "root:studio@tcp(127.0.0.1:3307)/huhnlite?parseTime=true&allowNativePasswords=true"
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening MySQL DB: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()

	// Let's test for Silo ID 2
	siloID := int64(2)
	inventurDatum := "2026-05-30"
	inventurMenge := 5000.0

	fmt.Printf("--- Dry Run Futterinventur for MySQL Silo ID: %d, Date: %s, Qty: %.2f ---\n", siloID, inventurDatum, inventurMenge)

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	// 1. Silo laden
	var silonummer int64
	var bezeichnung string
	var inventurDatumAlt string
	var inventurDatumNeu string
	err = tx.QueryRowContext(ctx, "SELECT SILONUMMER, BEZEICHNUNG, INVENTURDATUMALT, INVENTURDATUMNEU FROM SILO WHERE ID = ?", siloID).Scan(&silonummer, &bezeichnung, &inventurDatumAlt, &inventurDatumNeu)
	if err != nil {
		log.Fatalf("Error loading Silo: %v", err)
	}
	fmt.Printf("Silo Details: Nr=%d, Bez=%s, Alt=%s, Neu=%s\n", silonummer, bezeichnung, inventurDatumAlt, inventurDatumNeu)

	if strings.TrimSpace(inventurDatumAlt) == "" {
		inventurDatumAlt = "0001-01-01"
	}

	// 2. Check 1: Max(Buchungsdatum) > inventurDatum
	var maxBuchungsdatum sql.NullString
	err = tx.QueryRowContext(ctx, "SELECT MAX(BUCHUNGSDATUM) FROM BUCHUNG WHERE SILONR = ?", silonummer).Scan(&maxBuchungsdatum)
	if err != nil {
		log.Fatalf("Error query max buchung: %v", err)
	}
	maxDateStr := "0001-01-01"
	if maxBuchungsdatum.Valid {
		maxDateStr = maxBuchungsdatum.String
	}
	fmt.Printf("Max Buchungsdatum in BUCHUNG: %s\n", maxDateStr)

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
	fmt.Printf("Last Futter: ID_SILO=%d, LDate=%s, LQty=%.2f, PreisDt=%.2f, Rabatt=%.2f, MwStProz=%.2f, MwStKz=%v\n",
		lastF.ID_SILO, lastF.LIEFERDATUM, lastF.LIEFERMENGE, lastF.PREISDT, lastF.RABATTPROZ, lastF.MWSTPROZ, lastF.MWSTKZ)

	// 4. Calculate Netto & Brutto for A
	nettoA := inventurMenge * (lastF.PREISDT / 100.0) * (1.0 - lastF.RABATTPROZ/100.0)
	bruttoA := nettoA * (1.0 + lastF.MWSTPROZ/100.0)
	fmt.Printf("Booking A: Date=%s, Qty=%.2f, Net=%.2f, Brutto=%.2f\n", inventurDatum, inventurMenge, nettoA, bruttoA)

	insertQuery := `
		INSERT INTO FUTTER (ID_SILO, SILONUMMER, HERDENR, ID_PERSON, LIEFERDATUM, LIEFERMENGE, PREISDT, RABATTPROZ, NETTO, BRUTTO, MWSTPROZ, MWSTKZ, DATUM, ZEITSTEMPEL, AW, ID_FUTTERSORTEN)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	resA, err := tx.ExecContext(ctx, insertQuery,
		lastF.ID_SILO, lastF.SILONUMMER, lastF.HERDENR, lastF.ID_PERSON,
		inventurDatum, inventurMenge, lastF.PREISDT, lastF.RABATTPROZ,
		nettoA, bruttoA, lastF.MWSTPROZ, lastF.MWSTKZ,
		inventurDatum, inventurDatum+" 12:00:00Z", lastF.AW, lastF.ID_FUTTERSORTEN,
	)
	if err != nil {
		log.Fatalf("Error insert A: %v", err)
	}
	idA, _ := resA.LastInsertId()
	fmt.Printf("Successfully inserted booking A. LastInsertId=%d\n", idA)

	// 5. Booking B
	tParsed, err := time.Parse("2006-01-02", inventurDatum)
	if err != nil {
		log.Fatal(err)
	}
	prevDatum := tParsed.AddDate(0, 0, -1).Format("2006-01-02")
	mengeB := -inventurMenge
	nettoB := mengeB * (lastF.PREISDT / 100.0) * (1.0 - lastF.RABATTPROZ/100.0)
	bruttoB := nettoB * (1.0 + lastF.MWSTPROZ/100.0)
	fmt.Printf("Booking B: Date=%s, Qty=%.2f, Net=%.2f, Brutto=%.2f\n", prevDatum, mengeB, nettoB, bruttoB)

	resB, err := tx.ExecContext(ctx, insertQuery,
		lastF.ID_SILO, lastF.SILONUMMER, lastF.HERDENR, lastF.ID_PERSON,
		prevDatum, mengeB, lastF.PREISDT, lastF.RABATTPROZ,
		nettoB, bruttoB, lastF.MWSTPROZ, lastF.MWSTKZ,
		prevDatum, prevDatum+" 12:00:00Z", lastF.AW, lastF.ID_FUTTERSORTEN,
	)
	if err != nil {
		log.Fatalf("Error insert B: %v", err)
	}
	idB, _ := resB.LastInsertId()
	fmt.Printf("Successfully inserted booking B. LastInsertId=%d\n", idB)

	// 6. Period summation
	var gesamtLiefermenge, gesamtNetto, gesamtBrutto float64
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(LIEFERMENGE), 0.0), COALESCE(SUM(NETTO), 0.0), COALESCE(SUM(BRUTTO), 0.0)
		FROM FUTTER WHERE ID_SILO = ? AND LIEFERDATUM >= ? AND LIEFERDATUM <= ?`,
		siloID, inventurDatumAlt, prevDatum).Scan(&gesamtLiefermenge, &gesamtNetto, &gesamtBrutto)
	if err != nil {
		log.Fatalf("Error summation: %v", err)
	}
	fmt.Printf("Summation for period [%s to %s]: Qty=%.2f, Net=%.2f, Brutto=%.2f\n", inventurDatumAlt, prevDatum, gesamtLiefermenge, gesamtNetto, gesamtBrutto)

	// 7. Futtertage
	var futtertageGesamt int64
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(TIERBESTAND), 0) FROM BUCHUNG
		WHERE SILONR = ? AND BUCHUNGSDATUM >= ? AND BUCHUNGSDATUM <= ?`,
		silonummer, inventurDatumAlt, prevDatum).Scan(&futtertageGesamt)
	if err != nil {
		log.Fatalf("Error query bird days: %v", err)
	}
	fmt.Printf("Bird Days: %d\n", futtertageGesamt)

	if futtertageGesamt > 0 {
		futterverbrauchTierGrams := int64(math.Round((gesamtLiefermenge * 1000.0) / float64(futtertageGesamt)))
		futterkostenTier := gesamtNetto / float64(futtertageGesamt)
		fmt.Printf("Recalculated metrics: Consumption/Bird=%d g, Cost/Bird=%.6f EUR\n", futterverbrauchTierGrams, futterkostenTier)
	}

	fmt.Println("MySQL Dry-run completed successfully without committing.")
}
