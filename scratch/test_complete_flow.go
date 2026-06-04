package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	_ "github.com/glebarez/go-sqlite"
	db "huhnlite-wails/backend/db/repo"
)

func main() {
	dbPath := "C:/Users/hofma/GolandProjects/HuhnLite-Wails/HuhnLite.db"
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Error opening DB: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	siloID := int64(2)
	inventurDatum := "2026-05-30"
	inventurMenge := 5000.0

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

	if strings.TrimSpace(inventurDatumAlt) == "" {
		inventurDatumAlt = "0001-01-01"
	}

	// 2. Letzte Lieferungs-Info
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

	// 3. Insert A
	nettoA := inventurMenge * (lastF.PREISDT / 100.0) * (1.0 - lastF.RABATTPROZ/100.0)
	bruttoA := nettoA * (1.0 + lastF.MWSTPROZ/100.0)
	insertQuery := `
		INSERT INTO FUTTER (ID_SILO, SILONUMMER, HERDENR, ID_PERSON, LIEFERDATUM, LIEFERMENGE, PREISDT, RABATTPROZ, NETTO, BRUTTO, MWSTPROZ, MWSTKZ, DATUM, ZEITSTEMPEL, AW, ID_FUTTERSORTEN)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, insertQuery,
		lastF.ID_SILO, lastF.SILONUMMER, lastF.HERDENR, lastF.ID_PERSON,
		inventurDatum, inventurMenge, lastF.PREISDT, lastF.RABATTPROZ,
		nettoA, bruttoA, lastF.MWSTPROZ, lastF.MWSTKZ,
		inventurDatum, inventurDatum+" 12:00:00Z", lastF.AW, lastF.ID_FUTTERSORTEN,
	)
	if err != nil {
		log.Fatalf("Error insert A: %v", err)
	}

	// 4. Insert B
	tParsed, _ := time.Parse("2006-01-02", inventurDatum)
	prevDatum := tParsed.AddDate(0, 0, -1).Format("2006-01-02")
	mengeB := -inventurMenge
	nettoB := mengeB * (lastF.PREISDT / 100.0) * (1.0 - lastF.RABATTPROZ/100.0)
	bruttoB := nettoB * (1.0 + lastF.MWSTPROZ/100.0)

	_, err = tx.ExecContext(ctx, insertQuery,
		lastF.ID_SILO, lastF.SILONUMMER, lastF.HERDENR, lastF.ID_PERSON,
		prevDatum, mengeB, lastF.PREISDT, lastF.RABATTPROZ,
		nettoB, bruttoB, lastF.MWSTPROZ, lastF.MWSTKZ,
		prevDatum, prevDatum+" 12:00:00Z", lastF.AW, lastF.ID_FUTTERSORTEN,
	)
	if err != nil {
		log.Fatalf("Error insert B: %v", err)
	}

	// Summation
	var gesamtLiefermenge, gesamtNetto, gesamtBrutto float64
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(LIEFERMENGE), 0.0), COALESCE(SUM(NETTO), 0.0), COALESCE(SUM(BRUTTO), 0.0)
		FROM FUTTER WHERE ID_SILO = ? AND LIEFERDATUM >= ? AND LIEFERDATUM <= ?`,
		siloID, inventurDatumAlt, prevDatum).Scan(&gesamtLiefermenge, &gesamtNetto, &gesamtBrutto)
	if err != nil {
		log.Fatalf("Error summation: %v", err)
	}

	var futtertageGesamt int64
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(TIERBESTAND), 0) FROM BUCHUNG
		WHERE SILONR = ? AND BUCHUNGSDATUM >= ? AND BUCHUNGSDATUM <= ?`,
		silonummer, inventurDatumAlt, prevDatum).Scan(&futtertageGesamt)
	if err != nil {
		log.Fatalf("Error query bird days: %v", err)
	}

	if futtertageGesamt > 0 {
		futterverbrauchTierGrams := int64(math.Round((gesamtLiefermenge * 1000.0) / float64(futtertageGesamt)))
		futterkostenTier := gesamtNetto / float64(futtertageGesamt)

		// Update bookings
		_, err = tx.ExecContext(ctx, `
			UPDATE BUCHUNG
			SET FUTTERVERBRAUCHTIER = ?, FUTTERKTAG = ? * TIERBESTAND
			WHERE SILONR = ? AND BUCHUNGSDATUM >= ? AND BUCHUNGSDATUM <= ?`,
			futterverbrauchTierGrams, futterkostenTier, silonummer, inventurDatumAlt, prevDatum)
		if err != nil {
			log.Fatalf("Error updating bookings: %v", err)
		}
	}

	// Update Silo
	_, err = tx.ExecContext(ctx, `
		UPDATE SILO
		SET INVENTURDATUMALT = ?, INVENTURDATUMNEU = ?, INVENTURFUELLMENGE = ?
		WHERE ID = ?`,
		inventurDatum, inventurDatum, inventurMenge, siloID)
	if err != nil {
		log.Fatalf("Error updating Silo: %v", err)
	}

	// Commit transaction!
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Transaction committed successfully.")

	// Query bookings list using repo
	queries := db.New(conn)
	res, err := queries.ListFutterBuchungen(ctx)
	if err != nil {
		log.Fatalf("Error fetching list: %v", err)
	}
	fmt.Printf("Loaded %d bookings successfully.\n", len(res))

	// Clean up by deleting the test inserts
	_, err = conn.Exec("DELETE FROM FUTTER WHERE ID >= 1539")
	if err != nil {
		log.Fatalf("Error cleaning up: %v", err)
	}
	fmt.Println("Cleaned up database inserts.")
}
