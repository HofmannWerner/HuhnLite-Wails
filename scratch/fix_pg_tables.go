package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

func copyTableFromMySQLToPG(srcMySQLTable string, dstPGTable string, dstDBs []string) error {
	dsnMySQL := "root:studio@tcp(127.0.0.1:3307)/huhnlite_test?parseTime=true"
	dbMySQL, err := sql.Open("mysql", dsnMySQL)
	if err != nil {
		return err
	}
	defer dbMySQL.Close()

	rows, err := dbMySQL.Query(fmt.Sprintf("SELECT * FROM `%s`", srcMySQLTable))
	if err != nil {
		return fmt.Errorf("error querying MySQL table %s: %v", srcMySQLTable, err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	pgTableName := strings.ToLower(dstPGTable)

	for _, dstDbName := range dstDBs {
		fmt.Printf("Syncing %s -> Postgres DB %s (%s)...\n", srcMySQLTable, dstDbName, pgTableName)
		dsnPG := fmt.Sprintf("postgres://postgres:post@192.168.178.28:5432/%s?sslmode=disable", dstDbName)
		dbPG, err := sql.Open("postgres", dsnPG)
		if err != nil {
			log.Printf("PG Open error for %s: %v", dstDbName, err)
			continue
		}

		// Truncate PG table
		_, err = dbPG.Exec(fmt.Sprintf("TRUNCATE TABLE \"%s\" RESTART IDENTITY CASCADE", pgTableName))
		if err != nil {
			log.Printf("Error truncating PG table %s: %v", pgTableName, err)
		}

		// Build insert query
		placeholders := make([]string, len(cols))
		colNames := make([]string, len(cols))
		for i, c := range cols {
			colNames[i] = fmt.Sprintf("\"%s\"", strings.ToLower(c))
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}

		insertSQL := fmt.Sprintf("INSERT INTO \"%s\" (%s) VALUES (%s)",
			pgTableName,
			strings.Join(colNames, ", "),
			strings.Join(placeholders, ", "))

		// Re-query rows for each destination DB
		rowsSub, err := dbMySQL.Query(fmt.Sprintf("SELECT * FROM `%s`", srcMySQLTable))
		if err != nil {
			dbPG.Close()
			continue
		}

		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))

		count := 0
		for rowsSub.Next() {
			for i := range cols {
				valuePtrs[i] = &values[i]
			}
			rowsSub.Scan(valuePtrs...)

			// Convert []byte to string for string fields if needed
			for i, v := range values {
				if b, ok := v.([]byte); ok {
					values[i] = string(b)
				}
			}

			_, errExec := dbPG.Exec(insertSQL, values...)
			if errExec != nil {
				log.Printf("Insert error into %s: %v", pgTableName, errExec)
			} else {
				count++
			}
		}
		rowsSub.Close()
		dbPG.Close()
		fmt.Printf("  Successfully inserted %d rows into %s (%s)\n", count, dstDbName, pgTableName)
	}
	return nil
}

func main() {
	dstDBs := []string{"huhnlite-prod-1", "huhnlite-test-1", "huhnlite-prod", "huhnlite-test"}

	tablesToSync := []string{
		"DYNAMISCHE_SQL",
		"TEXTE",
		"TEXT_TYPEN",
		"UEBERSETZUNGEN",
		"TRANSLATEFELDNAMEN",
		"FELD_KATALOG",
	}

	for _, tbl := range tablesToSync {
		if err := copyTableFromMySQLToPG(tbl, tbl, dstDBs); err != nil {
			log.Printf("Error syncing %s: %v", tbl, err)
		}
	}

	fmt.Println("\n=== Table sync to Postgres complete! ===")
}
