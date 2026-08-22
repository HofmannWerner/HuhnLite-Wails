package main

import (
	"database/sql"
	"fmt"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

func checkReports(name, engine, connStr string) {
	fmt.Printf("\n=== Reports in %s (%s) ===\n", name, engine)
	var db *sql.DB
	var err error
	if engine == "sqlite" {
		db, err = sql.Open("sqlite", connStr)
	} else if engine == "mysql" {
		db, err = sql.Open("mysql", connStr)
	} else {
		db, err = sql.Open("postgres", connStr)
	}
	if err != nil {
		fmt.Printf("Open error: %v\n", err)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, beschreibung, sqlstatement FROM DYNAMISCHE_SQL LIMIT 5")
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var beschreibung, sqlstmt string
		rows.Scan(&id, &beschreibung, &sqlstmt)
		if len(sqlstmt) > 60 {
			sqlstmt = sqlstmt[:60] + "..."
		}
		fmt.Printf("  - [%d] %s => %q\n", id, beschreibung, sqlstmt)
	}
}

func main() {
	checkReports("SQLite HuhnLite.db", "sqlite", "HuhnLite.db")
	checkReports("MySQL huhnlite_test", "mysql", "root:studio@tcp(127.0.0.1:3307)/huhnlite_test?parseTime=true")
	checkReports("Postgres huhnlite-prod-1", "postgres", "postgres://postgres:post@192.168.178.28:5432/huhnlite-prod-1?sslmode=disable")
}
