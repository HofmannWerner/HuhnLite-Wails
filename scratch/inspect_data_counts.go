package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

func inspectMySQLDb(dbName string) {
	fmt.Printf("\n=== MySQL DB: %s ===\n", dbName)
	dsn := fmt.Sprintf("root:studio@tcp(127.0.0.1:3307)/%s?parseTime=true", dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("Error open: %v\n", err)
		return
	}
	defer db.Close()

	tables := []string{"HERDEN", "FUTTER", "BUCHUNG", "EILAGER", "FIRMENPARAMETER"}
	for _, t := range tables {
		var count int
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM `%s`", t)).Scan(&count)
		if err != nil {
			fmt.Printf("  Table %s: error %v\n", t, err)
		} else {
			fmt.Printf("  Table %s: %d rows\n", t, count)
		}
	}
}

func inspectPGDb(dbName string) {
	fmt.Printf("\n=== Postgres DB: %s ===\n", dbName)
	dsn := fmt.Sprintf("postgres://postgres:post@192.168.178.28:5432/%s?sslmode=disable", dbName)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Printf("Error open: %v\n", err)
		return
	}
	defer db.Close()

	tables := []string{"herden", "futter", "buchung", "eilager", "firmenparameter"}
	for _, t := range tables {
		var count int
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM \"%s\"", t)).Scan(&count)
		if err != nil {
			fmt.Printf("  Table %s: error %v\n", t, err)
		} else {
			fmt.Printf("  Table %s: %d rows\n", t, count)
		}
	}
}

func main() {
	mysqlDbs := []string{"huhnlite", "huhnlite-1", "huhnlite-prod-1", "huhnlite-test-1", "huhnlite_prod", "huhnlite_test"}
	for _, d := range mysqlDbs {
		inspectMySQLDb(d)
	}

	pgDbs := []string{"huhnlite-prod", "huhnlite-test", "huhnlite-prod-1", "huhnlite-test-1"}
	for _, d := range pgDbs {
		inspectPGDb(d)
	}
}
