package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
	"huhnlite-wails/backend/db/repo"
)

func main() {
	dbConn, err := sql.Open("sqlite", "HuhnLite.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer dbConn.Close()

	queries := repo.New(dbConn)

	params := repo.ListAktionenParams{
		IDUser:    int64(0),
		StartDate: "",
		EndDate:   "",
		Kz:        "",
		Status:    int64(0),
	}

	fmt.Printf("Running ListAktionen with params: %+v\n", params)

	res, err := queries.ListAktionen(context.Background(), params)
	if err != nil {
		log.Fatalf("ListAktionen failed: %v", err)
	}

	fmt.Printf("Success! Returned %d rows\n", len(res))
	for idx, row := range res {
		if idx >= 5 {
			fmt.Println("...")
			break
		}
		fmt.Printf("- Row %d: ID=%d, Kz=%v, Datum=%v, Bezeichnung=%v, Erledigt=%v\n",
			idx, row.ID, row.AktionenKz, row.Aktionsdatum.String, row.Bezeichnung.String, row.Erledigt.Int64)
	}
}
