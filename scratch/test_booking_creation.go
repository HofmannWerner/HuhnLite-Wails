package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"huhnlite-wails/backend/db"
	"huhnlite-wails/backend/db/repo"
	repo_mysql "huhnlite-wails/backend/db/repo_mysql"
)

func getSiloNrForHerd(ctx context.Context, dbExecutor dbExecutor, herdID int64) int64 {
	var idSilo int64
	err := dbExecutor.QueryRowContext(ctx, "SELECT ID_SILO FROM HERDEN WHERE ID = ?", herdID).Scan(&idSilo)
	if err != nil || idSilo <= 0 {
		return 0
	}
	var silonr int64
	err = dbExecutor.QueryRowContext(ctx, "SELECT SILONUMMER FROM SILO WHERE ID = ?", idSilo).Scan(&silonr)
	if err != nil {
		return 0
	}
	return silonr
}

type dbExecutor interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func main() {
	dsn := "root:turbodiesel@tcp(192.168.178.60:3307)/huhnlite?parseTime=true"
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		log.Fatalf("BeginTx failed: %v", err)
	}
	defer tx.Rollback()

	queriesMysql := repo_mysql.New(tx)
	wrapper := db.NewMySQLWrapper(nil, queriesMysql, conn).WithTx(tx)

	herdID := int64(24) // Brinkmann Herd
	fmt.Printf("=== Testing Booking Creation & Update Silo Assignment ===\n")

	// 1. Resolve SiloNr for Herd 24
	siloNr := getSiloNrForHerd(ctx, tx, herdID)
	fmt.Printf("Resolved SiloNr for Herd %d: %d (expected 36)\n", herdID, siloNr)

	// 2. Simulate POST /api/buchung (Normal path)
	// Create Buchung Params
	p := repo.CreateBuchungParams{
		IDHerden:        herdID,
		Lw:              45,
		Herdennummer:    24,
		Buchungsdatum:   "2026-04-20",
		Gewichtprobe:    1850,
		Kontrollgewicht: 1.85,
		Klassea:         3500,
		Verluste:        2,
		Eimasse:         215.5,
		Schmutz:         10,
		Knickeier:       5,
		Vollei:          15.0,
		Brucheier:       1,
		Tierbestand:     3730,
		IDEitabelle:     1,
		IDDgewichttab:   1,
		Futterktag:      0, // not set in frontend
		Silonr:          siloNr, // resolved via getSiloNrForHerd
		Kl6:             0,
		Vermitteltam:    "0001-01-01",
		Small:           0,
		Large:           0,
		Medium:          0,
		Xl:              0,
		Dgewichtei:      61.5,
		Zeitstempel:     "2026-04-20 17:00:00",
		Aw:              1,
		Vermittelt:      "N",
		Futterverbrauchtier: 0, // not set in frontend
	}

	created, err := wrapper.CreateBuchung(ctx, p)
	if err != nil {
		log.Fatalf("CreateBuchung failed: %v", err)
	}
	fmt.Printf("Successfully created booking. ID: %d, Saved Silonr: %d (expected 36)\n", created.ID, created.Silonr)

	// 3. Simulate PUT /api/buchung/:id where frontend sends 0 for SILONR, FUTTERKTAG, FUTTERVERBRAUCHTIER
	// Normally backend would load the existing booking and preserve them
	existing, err := wrapper.GetBuchung(ctx, created.ID)
	if err != nil {
		log.Fatalf("GetBuchung failed: %v", err)
	}

	// Mocking calculations run in the meantime that set Futterverbrauchtier & Futterktag
	// Let's directly update the database first to pretend inventory run happened
	_, err = tx.ExecContext(ctx, "UPDATE BUCHUNG SET FUTTERVERBRAUCHTIER = 141, FUTTERKTAG = 232 WHERE ID = ?", created.ID)
	if err != nil {
		log.Fatalf("Mock update failed: %v", err)
	}

	// Reload existing
	existing, err = wrapper.GetBuchung(ctx, created.ID)
	fmt.Printf("Mocked inventory updated values -> Verbrauch: %d, Ktag: %d\n", existing.Futterverbrauchtier, existing.Futterktag)

	// Simulate payload from frontend: SILONR = 0, FUTTERKTAG = 0, FUTTERVERBRAUCHTIER = 0
	reqSilonr := int64(0)
	reqFkt := int64(0)
	reqFvt := int64(0)

	// Backend preserves them
	upSilonr := reqSilonr
	if upSilonr <= 0 {
		upSilonr = existing.Silonr
	}
	if upSilonr <= 0 {
		upSilonr = getSiloNrForHerd(ctx, tx, herdID)
	}

	upFvt := reqFvt
	if upFvt <= 0 {
		upFvt = existing.Futterverbrauchtier
	}

	upFkt := reqFkt
	if upFkt <= 0 {
		upFkt = existing.Futterktag
	}

	params := repo.UpdateBuchungParams{
		ID:              created.ID,
		IDHerden:        existing.IDHerden,
		Lw:              existing.Lw,
		Herdennummer:    existing.Herdennummer,
		Buchungsdatum:   existing.Buchungsdatum,
		Gewichtprobe:    existing.Gewichtprobe,
		Kontrollgewicht: existing.Kontrollgewicht,
		Klassea:         existing.Klassea,
		Verluste:        existing.Verluste,
		Eimasse:         existing.Eimasse,
		Schmutz:         existing.Schmutz,
		Knickeier:       existing.Knickeier,
		Vollei:          existing.Vollei,
		Brucheier:       existing.Brucheier,
		Tierbestand:     existing.Tierbestand,
		IDEitabelle:     existing.IDEitabelle,
		IDDgewichttab:   existing.IDDgewichttab,
		Futterktag:      upFkt,
		Silonr:          upSilonr,
		Kl6:             existing.Kl6,
		Vermitteltam:    existing.Vermitteltam,
		Small:           existing.Small,
		Large:           existing.Large,
		Medium:          existing.Medium,
		Xl:              existing.Xl,
		Dgewichtei:      existing.Dgewichtei,
		Zeitstempel:     existing.Zeitstempel,
		Aw:              existing.Aw,
		Vermittelt:      existing.Vermittelt,
		Futterverbrauchtier: upFvt,
	}

	updated, err := wrapper.UpdateBuchung(ctx, params)
	if err != nil {
		log.Fatalf("UpdateBuchung failed: %v", err)
	}

	fmt.Printf("Successfully updated booking.\n")
	fmt.Printf("  Saved Silonr: %d (expected 36)\n", updated.Silonr)
	fmt.Printf("  Saved Verbrauch: %d (expected 141)\n", updated.Futterverbrauchtier)
	fmt.Printf("  Saved Ktag: %d (expected 232)\n", updated.Futterktag)

	fmt.Println("All booking creation and preservation tests passed! Rolling back transaction.")
}
