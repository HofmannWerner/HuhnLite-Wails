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

	fmt.Println("=== Describing BUCHUNG table in MySQL ===")
	rows, err := db.Query("DESCRIBE BUCHUNG")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var field, typ, null, key, extra string
		var defaultVal sql.NullString
		rows.Scan(&field, &typ, &null, &key, &defaultVal, &extra)
		fmt.Printf("- Field: %s | Type: %s | Null: %s | Key: %s | Default: %v | Extra: %s\n",
			field, typ, null, key, defaultVal, extra)
	}
}
