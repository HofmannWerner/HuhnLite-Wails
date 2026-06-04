package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func columnExists(db *sql.DB, tableName, columnName string) (bool, error) {
	query := fmt.Sprintf("SHOW COLUMNS FROM %s LIKE '%s'", tableName, columnName)
	rows, err := db.Query(query)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

func main() {
	dsn := "root:studio@tcp(127.0.0.1:3307)/huhnlite?parseTime=true&allowNativePasswords=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening MySQL DB: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Cannot connect to MySQL: %v", err)
	}
	fmt.Println("Connected to MySQL database successfully.")

	// 1. Check BUCHUNG.FUTTERVERBRAUCHTIER
	exists, err := columnExists(db, "BUCHUNG", "FUTTERVERBRAUCHTIER")
	if err != nil {
		log.Fatalf("Error checking column: %v", err)
	}
	if !exists {
		fmt.Println("Adding FUTTERVERBRAUCHTIER to BUCHUNG table...")
		_, err = db.Exec("ALTER TABLE BUCHUNG ADD COLUMN FUTTERVERBRAUCHTIER INT NOT NULL DEFAULT 0")
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Println("✅ FUTTERVERBRAUCHTIER successfully added to BUCHUNG!")
	} else {
		fmt.Println("FUTTERVERBRAUCHTIER column already exists in BUCHUNG.")
	}

	// 2. Check FIRMENPARAMETER.FUTTERINVENTUR
	exists, err = columnExists(db, "FIRMENPARAMETER", "FUTTERINVENTUR")
	if err != nil {
		log.Fatalf("Error checking column: %v", err)
	}
	if !exists {
		fmt.Println("Adding FUTTERINVENTUR to FIRMENPARAMETER table...")
		_, err = db.Exec("ALTER TABLE FIRMENPARAMETER ADD COLUMN FUTTERINVENTUR INT NOT NULL DEFAULT 0")
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Println("✅ FUTTERINVENTUR successfully added to FIRMENPARAMETER!")
	} else {
		fmt.Println("FUTTERINVENTUR column already exists in FIRMENPARAMETER.")
	}

	// 3. Check FUTTER.ZEITSTEMPEL
	exists, err = columnExists(db, "FUTTER", "ZEITSTEMPEL")
	if err != nil {
		log.Fatalf("Error checking column: %v", err)
	}
	if !exists {
		fmt.Println("Adding ZEITSTEMPEL to FUTTER table...")
		_, err = db.Exec("ALTER TABLE FUTTER ADD COLUMN ZEITSTEMPEL VARCHAR(50) DEFAULT '0001-01-01 00:00:00'")
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		fmt.Println("✅ ZEITSTEMPEL successfully added to FUTTER!")
	} else {
		fmt.Println("ZEITSTEMPEL column already exists in FUTTER.")
	}
}
