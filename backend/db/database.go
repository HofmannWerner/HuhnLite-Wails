package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"

	"huhnlite-wails/backend/config"
	"huhnlite-wails/backend/db/repo"
	"huhnlite-wails/backend/db/repo_mysql"
)

type DB struct {
	SQL       *sql.DB
	Repo      repo.Querier
	RepoMySQL *repo_mysql.Queries
	Config    config.Config
	Engine    string
}

func Connect(cfg config.Config) (*DB, error) {
	var conn *sql.DB
	var err error

	if cfg.DBEngine == "sqlite" {
		conn, err = sql.Open("sqlite", cfg.DBConnectionString)
	} else if cfg.DBEngine == "mysql" {
		// Expecting MariaDB/MySQL DSN
		conn, err = sql.Open("mysql", cfg.DBConnectionString)
	} else {
		return nil, fmt.Errorf("unsupported database engine: %s", cfg.DBEngine)
	}

	if err == nil {
		// Initialize schema if needed / Migrations (Shared for SQLite and MySQL)
		// Renaming SYSTEM to SYSTEM_KZ to avoid reserved word conflicts
		tables := []string{"TEXTE", "TEXT_TYPEN"}
		for _, table := range tables {
			var hasOldColumn bool
			if cfg.DBEngine == "sqlite" {
				// Check for old column in SQLite
				checkQuery := fmt.Sprintf("PRAGMA table_info(%s)", table)
				rows, err := conn.Query(checkQuery)
				if err == nil {
					for rows.Next() {
						var cid int
						var name, dtype string
						var notnull, pk int
						var dflt_value interface{}
						if err := rows.Scan(&cid, &name, &dtype, &notnull, &dflt_value, &pk); err == nil {
							if name == "SYSTEM" {
								hasOldColumn = true
							}
						}
					}
					rows.Close()
				}
			} else {
				// Check for old column in MariaDB/MySQL
				checkQuery := fmt.Sprintf("SHOW COLUMNS FROM %s LIKE 'SYSTEM'", table)
				var dummy string
				err := conn.QueryRow(checkQuery).Scan(&dummy, &dummy, &dummy, &dummy, &dummy, &dummy)
				if err == nil {
					hasOldColumn = true
				}
			}

			if hasOldColumn {
				// Check if SYSTEM_KZ also exists
				hasNewColumn := false
				if cfg.DBEngine == "sqlite" {
					if rows, err := conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", table)); err == nil {
						for rows.Next() {
							var cid int
							var name, dtype string
							var notnull, pk int
							var dflt_value interface{}
							if err := rows.Scan(&cid, &name, &dtype, &notnull, &dflt_value, &pk); err == nil && name == "SYSTEM_KZ" {
								hasNewColumn = true
							}
						}
						rows.Close()
					}
				} else {
					var dummy string
					err := conn.QueryRow(fmt.Sprintf("SHOW COLUMNS FROM %s LIKE 'SYSTEM_KZ'", table)).Scan(&dummy, &dummy, &dummy, &dummy, &dummy, &dummy)
					if err == nil {
						hasNewColumn = true
					}
				}

				if hasNewColumn {
					log.Printf("[DB] Both SYSTEM and SYSTEM_KZ exist in %s, dropping old SYSTEM column...", table)
					if _, err := conn.Exec(fmt.Sprintf("ALTER TABLE %s DROP COLUMN `SYSTEM`", table)); err == nil {
						log.Printf("[DB] Successfully dropped old SYSTEM column in %s", table)
					} else {
						log.Printf("[DB] Error dropping old SYSTEM column in %s: %v", table, err)
					}
				} else {
					log.Printf("[DB] Old SYSTEM column found in %s, renaming to SYSTEM_KZ...", table)
					if cfg.DBEngine == "sqlite" {
						if _, err := conn.Exec(fmt.Sprintf("ALTER TABLE %s RENAME COLUMN `SYSTEM` TO SYSTEM_KZ", table)); err == nil {
							log.Printf("[DB] Successfully renamed SYSTEM to SYSTEM_KZ in %s (sqlite)", table)
						} else {
							log.Printf("[DB] Error renaming SYSTEM to SYSTEM_KZ in %s: %v", table, err)
						}
					} else {
						if _, err := conn.Exec(fmt.Sprintf("ALTER TABLE %s CHANGE COLUMN `SYSTEM` SYSTEM_KZ INTEGER NOT NULL DEFAULT 0", table)); err == nil {
							log.Printf("[DB] Successfully renamed SYSTEM to SYSTEM_KZ in %s (mysql)", table)
						} else {
							log.Printf("[DB] Error renaming SYSTEM to SYSTEM_KZ in %s: %v", table, err)
						}
					}
				}
			} else {
				// Try to add the new column SYSTEM_KZ if it doesn't exist
				if _, err := conn.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN SYSTEM_KZ INTEGER NOT NULL DEFAULT 0", table)); err == nil {
					log.Printf("[DB] Successfully added SYSTEM_KZ column to %s (%s)", table, cfg.DBEngine)
				}
			}
		}
	} else {
		return nil, fmt.Errorf("unsupported database engine: %s", cfg.DBEngine)
	}

	if err == nil {
		// Fix für AKTIONEN (ERLEDIGT_AM und ID_USER_ERLEDIGT hinzufügen falls fehlend)
		// Muss AUSSERHALB der TEXTE-Schleife liegen
		{
			table := "AKTIONEN"
			columns := []struct {
				name string
				spec string
			}{
				{"ERLEDIGT_AM", "VARCHAR(25) DEFAULT ''"},
				{"ID_USER_ERLEDIGT", "INTEGER DEFAULT 0"},
			}
			for _, col := range columns {
				var hasColumn bool
				if cfg.DBEngine == "sqlite" {
					if rows, err := conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", table)); err == nil {
						for rows.Next() {
							var cid int
							var name, dtype string
							var notnull, pk int
							var dflt_value interface{}
							if err := rows.Scan(&cid, &name, &dtype, &notnull, &dflt_value, &pk); err == nil && name == col.name {
								hasColumn = true
							}
						}
						rows.Close()
					}
				} else {
					checkQuery := fmt.Sprintf("SHOW COLUMNS FROM %s LIKE '%s'", table, col.name)
					// Scan columns (Field, Type, Null, Key, Default, Extra)
					rows, err := conn.Query(checkQuery)
					if err == nil {
						if rows.Next() {
							hasColumn = true
						}
						rows.Close()
					}
				}
				if !hasColumn {
					log.Printf("[DB] Adding missing column %s to %s (%s)...", col.name, table, cfg.DBEngine)
					spec := col.spec
					if cfg.DBEngine == "sqlite" && strings.Contains(spec, "VARCHAR") {
						spec = "TEXT DEFAULT ''"
					}
					if _, err := conn.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, col.name, spec)); err == nil {
						log.Printf("[DB] Successfully added %s to %s", col.name, table)
					} else {
						log.Printf("[DB] Error adding %s to %s: %v", col.name, table, err)
					}
				}
			}
		}

		// Fix für FUTTER (ZEITSTEMPEL hinzufügen falls fehlend)
		{
			table := "FUTTER"
			columns := []struct {
				name string
				spec string
			}{
				{"ZEITSTEMPEL", "VARCHAR(50) DEFAULT '0001-01-01 00:00:00'"},
			}
			for _, col := range columns {
				var hasColumn bool
				if cfg.DBEngine == "sqlite" {
					if rows, err := conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", table)); err == nil {
						for rows.Next() {
							var cid int
							var name, dtype string
							var notnull, pk int
							var dflt_value interface{}
							if err := rows.Scan(&cid, &name, &dtype, &notnull, &dflt_value, &pk); err == nil && name == col.name {
								hasColumn = true
							}
						}
						rows.Close()
					}
				} else {
					checkQuery := fmt.Sprintf("SHOW COLUMNS FROM %s LIKE '%s'", table, col.name)
					rows, err := conn.Query(checkQuery)
					if err == nil {
						if rows.Next() {
							hasColumn = true
						}
						rows.Close()
					}
				}
				if !hasColumn {
					log.Printf("[DB] Adding missing column %s to %s (%s)...", col.name, table, cfg.DBEngine)
					spec := col.spec
					if cfg.DBEngine == "sqlite" && strings.Contains(spec, "VARCHAR") {
						spec = "TEXT DEFAULT '0001-01-01 00:00:00'"
					}
					if _, err := conn.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, col.name, spec)); err == nil {
						log.Printf("[DB] Successfully added %s to %s", col.name, table)
					} else {
						log.Printf("[DB] Error adding %s to %s: %v", col.name, table, err)
					}
				}
			}
		}

		// Fix für BUCHUNG (FUTTERVERBRAUCHTIER hinzufügen falls fehlend)
		{
			table := "BUCHUNG"
			columns := []struct {
				name string
				spec string
			}{
				{"FUTTERVERBRAUCHTIER", "INTEGER NOT NULL DEFAULT 0"},
			}
			for _, col := range columns {
				var hasColumn bool
				if cfg.DBEngine == "sqlite" {
					if rows, err := conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", table)); err == nil {
						for rows.Next() {
							var cid int
							var name, dtype string
							var notnull, pk int
							var dflt_value interface{}
							if err := rows.Scan(&cid, &name, &dtype, &notnull, &dflt_value, &pk); err == nil && name == col.name {
								hasColumn = true
							}
						}
						rows.Close()
					}
				} else {
					checkQuery := fmt.Sprintf("SHOW COLUMNS FROM %s LIKE '%s'", table, col.name)
					rows, err := conn.Query(checkQuery)
					if err == nil {
						if rows.Next() {
							hasColumn = true
						}
						rows.Close()
					}
				}
				if !hasColumn {
					log.Printf("[DB] Adding missing column %s to %s (%s)...", col.name, table, cfg.DBEngine)
					spec := col.spec
					if cfg.DBEngine == "sqlite" && strings.Contains(spec, "INTEGER") {
						spec = "INTEGER NOT NULL DEFAULT 0"
					} else if cfg.DBEngine == "mysql" && strings.Contains(spec, "INTEGER") {
						spec = "INT NOT NULL DEFAULT 0"
					}
					if _, err := conn.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, col.name, spec)); err == nil {
						log.Printf("[DB] Successfully added %s to %s", col.name, table)
					} else {
						log.Printf("[DB] Error adding %s to %s: %v", col.name, table, err)
					}
				}
			}
		}

		// Fix für FIRMENPARAMETER (FUTTERINVENTUR hinzufügen falls fehlend)
		{
			table := "FIRMENPARAMETER"
			columns := []struct {
				name string
				spec string
			}{
				{"FUTTERINVENTUR", "INTEGER NOT NULL DEFAULT 0"},
			}
			for _, col := range columns {
				var hasColumn bool
				if cfg.DBEngine == "sqlite" {
					if rows, err := conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", table)); err == nil {
						for rows.Next() {
							var cid int
							var name, dtype string
							var notnull, pk int
							var dflt_value interface{}
							if err := rows.Scan(&cid, &name, &dtype, &notnull, &dflt_value, &pk); err == nil && name == col.name {
								hasColumn = true
							}
						}
						rows.Close()
					}
				} else {
					checkQuery := fmt.Sprintf("SHOW COLUMNS FROM %s LIKE '%s'", table, col.name)
					rows, err := conn.Query(checkQuery)
					if err == nil {
						if rows.Next() {
							hasColumn = true
						}
						rows.Close()
					}
				}
				if !hasColumn {
					log.Printf("[DB] Adding missing column %s to %s (%s)...", col.name, table, cfg.DBEngine)
					spec := col.spec
					if cfg.DBEngine == "sqlite" && strings.Contains(spec, "INTEGER") {
						spec = "INTEGER NOT NULL DEFAULT 0"
					} else if cfg.DBEngine == "mysql" && strings.Contains(spec, "INTEGER") {
						spec = "INT NOT NULL DEFAULT 0"
					}
					if _, err := conn.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, col.name, spec)); err == nil {
						log.Printf("[DB] Successfully added %s to %s", col.name, table)
					} else {
						log.Printf("[DB] Error adding %s to %s: %v", col.name, table, err)
					}
				}
			}
		}

		// Fix for existing float values in FUTTERKTAG in SQLite
		if cfg.DBEngine == "sqlite" {
			_, err = conn.Exec("UPDATE BUCHUNG SET FUTTERKTAG = CAST(ROUND(FUTTERKTAG) AS INTEGER) WHERE typeof(FUTTERKTAG) = 'real'")
			if err != nil {
				log.Printf("[DB] Error cleaning up FUTTERKTAG floats: %v", err)
			} else {
				log.Printf("[DB] Successfully cleaned up any float FUTTERKTAG values in SQLite")
			}
		}
	}

	if err != nil {
		return nil, err
	}

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	log.Printf("Successfully connected to %s database at %s", cfg.DBEngine, cfg.DBConnectionString)

	d := &DB{
		SQL:       conn,
		RepoMySQL: repo_mysql.New(conn),
		Config:    cfg,
		Engine:    cfg.DBEngine,
	}

	if cfg.DBEngine == "mysql" {
		d.Repo = NewMySQLWrapper(repo.New(conn), d.RepoMySQL, conn)
	} else {
		d.Repo = repo.New(conn)
	}

	return d, nil
}
