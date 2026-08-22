package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

func dumpAndCopyMySQL(srcDb, dstDb string) error {
	fmt.Printf("Copying MySQL database %s -> %s...\n", srcDb, dstDb)
	dsnRoot := "root:studio@tcp(127.0.0.1:3307)/?parseTime=true&multiStatements=true"
	db, err := sql.Open("mysql", dsnRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	// Create dstDb if not exists
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dstDb))
	if err != nil {
		return fmt.Errorf("error creating DB %s: %v", dstDb, err)
	}

	// Get tables from srcDb
	rows, err := db.Query(fmt.Sprintf("SHOW TABLES FROM `%s`", srcDb))
	if err != nil {
		return fmt.Errorf("error showing tables from %s: %v", srcDb, err)
	}
	var tables []string
	for rows.Next() {
		var tbl string
		rows.Scan(&tbl)
		tables = append(tables, tbl)
	}
	rows.Close()

	// Disable foreign key checks on dstDb
	db.Exec(fmt.Sprintf("USE `%s`", dstDb))
	db.Exec("SET FOREIGN_KEY_CHECKS = 0")

	for _, tbl := range tables {
		// Drop dst table
		_, _ = db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`.`%s`", dstDb, tbl))
		// Create table like src table
		_, err := db.Exec(fmt.Sprintf("CREATE TABLE `%s`.`%s` LIKE `%s`.`%s`", dstDb, tbl, srcDb, tbl))
		if err != nil {
			log.Printf("  Error creating table %s in %s: %v", tbl, dstDb, err)
			continue
		}
		// Copy data
		_, err = db.Exec(fmt.Sprintf("INSERT INTO `%s`.`%s` SELECT * FROM `%s`.`%s`", dstDb, tbl, srcDb, tbl))
		if err != nil {
			log.Printf("  Error copying data for %s to %s: %v", tbl, dstDb, err)
		} else {
			fmt.Printf("  Successfully copied table %s to %s\n", tbl, dstDb)
		}
	}
	db.Exec("SET FOREIGN_KEY_CHECKS = 1")
	return nil
}

func dumpAndCopyPG(srcDb, dstDb string) error {
	fmt.Printf("Copying Postgres database %s -> %s...\n", srcDb, dstDb)
	dsnRoot := "postgres://postgres:post@192.168.178.28:5432/postgres?sslmode=disable"
	db, err := sql.Open("postgres", dsnRoot)
	if err != nil {
		return err
	}
	defer db.Close()

	// Drop and recreate dstDb
	db.Exec(fmt.Sprintf("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '%s'", dstDb))
	_, _ = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS \"%s\"", dstDb))
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE \"%s\" TEMPLATE \"%s\"", dstDb, srcDb))
	if err != nil {
		return fmt.Errorf("error creating PG database %s from template %s: %v", dstDb, srcDb, err)
	}
	fmt.Printf("  Successfully created PG DB %s from %s!\n", dstDb, srcDb)
	return nil
}

func main() {
	// 1. Copy MySQL huhnlite_test -> huhnlite-test-1 and huhnlite-prod-1
	if err := dumpAndCopyMySQL("huhnlite_test", "huhnlite-test-1"); err != nil {
		log.Printf("MySQL copy test-1 error: %v", err)
	}
	if err := dumpAndCopyMySQL("huhnlite_test", "huhnlite-prod-1"); err != nil {
		log.Printf("MySQL copy prod-1 error: %v", err)
	}

	// 2. Copy PG huhnlite-test-1 -> huhnlite-prod-1
	if err := dumpAndCopyPG("huhnlite-test-1", "huhnlite-prod-1"); err != nil {
		log.Printf("PG copy prod-1 error: %v", err)
	}

	fmt.Println("\n=== Database sync complete! ===")
}
