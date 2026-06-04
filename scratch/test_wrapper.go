package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"huhnlite-wails/backend/db"
	repo_mysql "huhnlite-wails/backend/db/repo_mysql"
)

func main() {
	dsn := "root:turbodiesel@tcp(192.168.178.60:3307)/huhnlite?parseTime=true"
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening MySQL DB: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()

	// Instantiate MySQLWrapper using NewMySQLWrapper helper
	queriesMysql := repo_mysql.New(conn)
	wrapper := db.NewMySQLWrapper(nil, queriesMysql, conn)

	// Let's find a booking with non-zero FUTTERKTAG
	var bookingID int64
	err = conn.QueryRow("SELECT COALESCE(MAX(ID), 0) FROM BUCHUNG WHERE FUTTERKTAG > 0").Scan(&bookingID)
	if err != nil {
		log.Fatalf("Failed to find non-zero booking: %v", err)
	}

	if bookingID == 0 {
		fmt.Println("No bookings with FUTTERKTAG > 0 found.")
		return
	}

	fmt.Printf("--- Fetching booking ID %d via MySQLWrapper ---\n", bookingID)

	b, err := wrapper.GetBuchung(ctx, bookingID)
	if err != nil {
		log.Fatalf("GetBuchung failed: %v", err)
	}

	fmt.Printf("Successfully loaded booking:\n")
	fmt.Printf("  ID: %d\n", b.ID)
	fmt.Printf("  Herd ID: %d\n", b.IDHerden)
	fmt.Printf("  Date: %s\n", b.Buchungsdatum)
	fmt.Printf("  Silo Nr: %d\n", b.Silonr)
	fmt.Printf("  Futterverbrauch Tier: %d\n", b.Futterverbrauchtier)
	fmt.Printf("  Futter Kosten Tag: %d\n", b.Futterktag)
}
