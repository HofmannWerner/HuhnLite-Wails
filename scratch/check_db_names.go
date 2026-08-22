package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

func testMySQL(host string) {
	fmt.Printf("\n--- Testing MySQL on %s ---\n", host)
	dsn := fmt.Sprintf("root:studio@tcp(%s)/?parseTime=true&timeout=3s", host)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("Open error: %v\n", err)
		return
	}
	defer db.Close()
	db.SetConnMaxLifetime(3 * time.Second)

	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		fmt.Printf("Query error on %s: %v\n", host, err)
		return
	}
	defer rows.Close()
	fmt.Printf("Databases on %s:\n", host)
	for rows.Next() {
		var name string
		rows.Scan(&name)
		fmt.Printf("  - %s\n", name)
	}
}

func testPostgres(host string) {
	fmt.Printf("\n--- Testing Postgres on %s ---\n", host)
	dsn := fmt.Sprintf("postgres://postgres:post@%s/postgres?sslmode=disable&connect_timeout=3", host)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Printf("Open error: %v\n", err)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT datname FROM pg_database WHERE datistemplate = false")
	if err != nil {
		fmt.Printf("Query error on %s: %v\n", host, err)
		return
	}
	defer rows.Close()
	fmt.Printf("Databases on %s:\n", host)
	for rows.Next() {
		var name string
		rows.Scan(&name)
		fmt.Printf("  - %s\n", name)
	}
}

func main() {
	testMySQL("127.0.0.1:3307")
	testMySQL("127.0.0.1:3306")
	testMySQL("192.168.178.60:3307")
	testMySQL("192.168.178.28:3307")

	testPostgres("127.0.0.1:5432")
	testPostgres("192.168.178.28:5432")
	testPostgres("192.168.178.60:5432")
}
