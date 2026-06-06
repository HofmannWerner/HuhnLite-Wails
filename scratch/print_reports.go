package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:studio@tcp(127.0.0.1:3307)/huhnlite?parseTime=true")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT ID, BESCHREIBUNG, SQLSTATEMENT, DETAIL_SQL, SUMMENZEILE FROM DYNAMISCHE_SQL WHERE TYP_KZ != 'H'")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var desc, sqlStmt, detailSql, summenzeile sql.NullString
		if err := rows.Scan(&id, &desc, &sqlStmt, &detailSql, &summenzeile); err == nil {
			fmt.Printf("\n--- REPORT %d: %s ---\n", id, desc.String)
			fmt.Printf("SQL:\n%s\n", sqlStmt.String)
			if detailSql.Valid && detailSql.String != "" {
				fmt.Printf("DETAIL_SQL:\n%s\n", detailSql.String)
			}
			if summenzeile.Valid && summenzeile.String != "" {
				fmt.Printf("SUMMENZEILE:\n%s\n", summenzeile.String)
			}
		}
	}
}
