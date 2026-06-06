package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 1. Migrate SQLite
	fmt.Println("=== Migrating SQLite ===")
	dbSqlite, err := sql.Open("sqlite", "C:/Users/hofma/AppData/Roaming/HuhnLite-Wails/HuhnLite.db")
	if err == nil {
		migrateDB(dbSqlite)
		dbSqlite.Close()
	} else {
		log.Printf("SQLite error: %v", err)
	}

	// Also migrate local project SQLite if it exists
	fmt.Println("\n=== Migrating SQLite in project root ===")
	dbProject, err := sql.Open("sqlite", "C:/Users/hofma/GolandProjects/HuhnLite-Wails/HuhnLite.db")
	if err == nil {
		migrateDB(dbProject)
		dbProject.Close()
	} else {
		log.Printf("Project SQLite error: %v", err)
	}

	// Also migrate build/bin SQLite if it exists
	fmt.Println("\n=== Migrating SQLite in build/bin ===")
	dbBin, err := sql.Open("sqlite", "C:/Users/hofma/GolandProjects/HuhnLite-Wails/build/bin/HuhnLite.db")
	if err == nil {
		migrateDB(dbBin)
		dbBin.Close()
	} else {
		log.Printf("Bin SQLite error: %v", err)
	}

	// 2. Migrate MySQL
	fmt.Println("\n=== Migrating MySQL ===")
	dbMysql, err := sql.Open("mysql", "root:studio@tcp(127.0.0.1:3307)/huhnlite?parseTime=true")
	if err == nil {
		migrateDB(dbMysql)
		dbMysql.Close()
	} else {
		log.Printf("MySQL error: %v", err)
	}
}

func migrateDB(db *sql.DB) {
	// Load all FELD_KATALOG records into a map
	fkMap := make(map[string]int64)
	rowsFK, err := db.Query("SELECT ID, FELDNAME FROM FELD_KATALOG")
	if err != nil {
		log.Printf("Error reading FELD_KATALOG: %v", err)
		return
	}
	for rowsFK.Next() {
		var id int64
		var name string
		rowsFK.Scan(&id, &name)
		fkMap[strings.ToUpper(strings.TrimSpace(name))] = id
	}
	rowsFK.Close()

	// Load all translation rows where ID_FELD_KATALOG is between 1 and 164
	type TransRow struct {
		OldFKID int64
		Lang    string
		Betreff string
		Inhalt  string
	}
	rowsT, err := db.Query(`
		SELECT ID_FELD_KATALOG, SPRACHE_KZ, BETREFF, INHALT 
		FROM TRANSLATEFELDNAMEN 
		WHERE ID_FELD_KATALOG BETWEEN 1 AND 164
	`)
	if err != nil {
		log.Printf("Error reading TRANSLATEFELDNAMEN: %v", err)
		return
	}
	var transRows []TransRow
	for rowsT.Next() {
		var r TransRow
		var betreff, inhalt sql.NullString
		rowsT.Scan(&r.OldFKID, &r.Lang, &betreff, &inhalt)
		r.Betreff = strings.TrimSpace(betreff.String)
		r.Inhalt = strings.TrimSpace(inhalt.String)
		transRows = append(transRows, r)
	}
	rowsT.Close()

	// Group German translations to find the matching FELDNAME
	germanNames := make(map[int64]string)
	for _, r := range transRows {
		if r.Lang == "de" && r.Betreff != "" {
			germanNames[r.OldFKID] = r.Betreff
		}
	}

	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return
	}
	defer tx.Rollback()

	insertedCount := 0
	updatedCount := 0
	deletedCount := 0

	for _, r := range transRows {
		germanName, hasGerman := germanNames[r.OldFKID]
		if !hasGerman {
			// Fallback: use the Betreff directly if it's the German source name
			germanName = r.Betreff
		}

		cleanGerman := strings.ToUpper(strings.TrimSpace(germanName))
		newID, found := fkMap[cleanGerman]
		if !found {
			// No active FELD_KATALOG match, keep it or skip it
			continue
		}

		// Check if a record for (newID, r.Lang) already exists
		var count int
		tx.QueryRow("SELECT COUNT(*) FROM TRANSLATEFELDNAMEN WHERE ID_FELD_KATALOG = ? AND SPRACHE_KZ = ?", newID, r.Lang).Scan(&count)

		if count > 0 {
			// Update existing translation
			_, err = tx.Exec(`
				UPDATE TRANSLATEFELDNAMEN 
				SET BETREFF = ?, INHALT = ? 
				WHERE ID_FELD_KATALOG = ? AND SPRACHE_KZ = ?`,
				r.Betreff, r.Inhalt, newID, r.Lang)
			if err == nil {
				updatedCount++
			} else {
				log.Printf("Update error: %v", err)
			}
		} else {
			// Insert new translation
			_, err = tx.Exec(`
				INSERT INTO TRANSLATEFELDNAMEN (ID_FELD_KATALOG, SPRACHE_KZ, BETREFF, INHALT) 
				VALUES (?, ?, ?, ?)`,
				newID, r.Lang, r.Betreff, r.Inhalt)
			if err == nil {
				insertedCount++
			} else {
				log.Printf("Insert error: %v", err)
			}
		}
	}

	// Delete old orphan records between 1 and 164
	res, err := tx.Exec("DELETE FROM TRANSLATEFELDNAMEN WHERE ID_FELD_KATALOG BETWEEN 1 AND 164")
	if err == nil {
		rowsDeleted, _ := res.RowsAffected()
		deletedCount = int(rowsDeleted)
	} else {
		log.Printf("Delete error: %v", err)
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Commit failed: %v", err)
		return
	}

	fmt.Printf("Migration completed: %d translations inserted, %d updated, %d orphan records cleaned up.\n",
		insertedCount, updatedCount, deletedCount)
}
