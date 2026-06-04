package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "C:/Users/hofma/GolandProjects/HuhnLite-Wails/HuhnLite.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Migrating FUTTERKTAG in BUCHUNG to DECIMAL(18,2) in SQLite...")

	// 1. Recreate table BUCHUNG to use DECIMAL(18, 2) / REAL for FUTTERKTAG
	// We rename current table
	_, err = db.Exec("ALTER TABLE BUCHUNG RENAME TO BUCHUNG_old")
	if err != nil {
		log.Fatalf("Error renaming table: %v", err)
	}

	// Create new BUCHUNG table
	createTableSQL := `CREATE TABLE BUCHUNG
(
    ID              INTEGER        NOT NULL,
    ID_HERDEN       INTEGER        NOT NULL,
    LW              INTEGER        NOT NULL DEFAULT 0,
    HERDENNUMMER    INTEGER        NOT NULL DEFAULT 0,
    BUCHUNGSDATUM   VARCHAR(25)    NOT NULL DEFAULT '0001-01-01',
    GEWICHTPROBE    INTEGER        NOT NULL DEFAULT 0,
    KONTROLLGEWICHT REAL           NOT NULL DEFAULT 0,
    KLASSEA         INTEGER        NOT NULL DEFAULT 0,
    VERLUSTE        INTEGER        NOT NULL DEFAULT 0,
    EIMASSE         REAL           NOT NULL DEFAULT 0,
    SCHMUTZ         INTEGER        NOT NULL DEFAULT 0,
    KNICKEIER       INTEGER        NOT NULL DEFAULT 0,
    VOLLEI          REAL           NOT NULL DEFAULT 0,
    BRUCHEIER       INTEGER        NOT NULL DEFAULT 0,
    TIERBESTAND     INTEGER        NOT NULL DEFAULT 0,
    ID_EITABELLE    INTEGER        NOT NULL DEFAULT 0,
    ID_DGEWICHTTAB  INTEGER        NOT NULL DEFAULT 0,
    FUTTERKTAG      DECIMAL(18, 2) NOT NULL DEFAULT 0,
    SILONR          INTEGER        NOT NULL DEFAULT 0,
    KL6             INTEGER        NOT NULL DEFAULT 0,
    VERMITTELTAM    VARCHAR(25)    NOT NULL DEFAULT '0001-01-01',
    SMALL           INTEGER        NOT NULL DEFAULT 0,
    LARGE           INTEGER        NOT NULL DEFAULT 0,
    MEDIUM          INTEGER        NOT NULL DEFAULT 0,
    XL              INTEGER        NOT NULL DEFAULT 0,
    ZEITSTEMPEL     VARCHAR(50)    NOT NULL DEFAULT '',
    DGEWICHTEI      DECIMAL(10, 2) NOT NULL DEFAULT 0,
    AW              INTEGER        NOT NULL DEFAULT 0,
    VERMITTELT      CHAR(1)        NOT NULL DEFAULT 'x',
PRIMARY KEY (ID)
);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Error creating new table: %v", err)
	}

	// Copy data with CAST to REAL/DECIMAL
	copySQL := `INSERT INTO BUCHUNG (
		ID, ID_HERDEN, LW, HERDENNUMMER, BUCHUNGSDATUM, GEWICHTPROBE, KONTROLLGEWICHT,
		KLASSEA, VERLUSTE, EIMASSE, SCHMUTZ, KNICKEIER, VOLLEI, BRUCHEIER, TIERBESTAND,
		ID_EITABELLE, ID_DGEWICHTTAB, FUTTERKTAG, SILONR, KL6, VERMITTELTAM, SMALL,
		LARGE, MEDIUM, XL, ZEITSTEMPEL, DGEWICHTEI, AW, VERMITTELT
	) SELECT 
		ID, ID_HERDEN, LW, HERDENNUMMER, BUCHUNGSDATUM, GEWICHTPROBE, KONTROLLGEWICHT,
		KLASSEA, VERLUSTE, EIMASSE, SCHMUTZ, KNICKEIER, VOLLEI, BRUCHEIER, TIERBESTAND,
		ID_EITABELLE, ID_DGEWICHTTAB, CAST(FUTTERKTAG AS REAL), SILONR, KL6, VERMITTELTAM, SMALL,
		LARGE, MEDIUM, XL, ZEITSTEMPEL, DGEWICHTEI, AW, VERMITTELT
	FROM BUCHUNG_old;`

	_, err = db.Exec(copySQL)
	if err != nil {
		log.Fatalf("Error copying data: %v", err)
	}

	// Drop old table
	_, err = db.Exec("DROP TABLE BUCHUNG_old")
	if err != nil {
		log.Fatalf("Error dropping old table: %v", err)
	}

	fmt.Println("SQLite database migration of BUCHUNG (FUTTERKTAG to DECIMAL) successfully completed!")
}
