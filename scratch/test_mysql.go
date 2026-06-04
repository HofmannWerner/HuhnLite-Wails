package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	db "huhnlite-wails/backend/db/repo_mysql"
)

func main() {
	dsn := "root:studio@tcp(127.0.0.1:3307)/huhnlite?parseTime=true&allowNativePasswords=true"
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening MySQL DB: %v", err)
	}
	defer conn.Close()

	err = conn.Ping()
	if err != nil {
		fmt.Printf("Cannot ping MySQL database: %v. MySQL might not be running or credentials mismatch.\n", err)
		return
	}
	fmt.Println("Ping successful. MySQL is running.")

	// Check columns of FUTTER
	fmt.Println("\n=== Columns in MySQL FUTTER ===")
	rows, err := conn.Query("DESCRIBE FUTTER")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var field, typ, null, key string
		var def, extra sql.NullString
		if err := rows.Scan(&field, &typ, &null, &key, &def, &extra); err == nil {
			fmt.Printf("- %s (%s, null: %s, key: %s, default: %s)\n", field, typ, null, key, def.String)
		}
	}

	// Query ListFutterBuchungen using the MySQL repo
	queries := db.New(conn)
	res, err := queries.ListFutterBuchungen(context.Background())
	if err != nil {
		fmt.Printf("\nERROR running ListFutterBuchungen on MySQL: %v\n", err)
		return
	}

	fmt.Printf("\nSuccessfully loaded %d feed bookings from MySQL.\n", len(res))
	if len(res) > 0 {
		fmt.Println("First few bookings:")
		for i := 0; i < len(res) && i < 5; i++ {
			fmt.Printf("- ID: %d, SiloNr: %d, Date: %s, Qty: %v, MwStKz: %v\n",
				res[i].ID, res[i].Silonummer, res[i].Lieferdatum, res[i].Liefermenge, res[i].Mwstkz)
		}
	}
}
