package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"huhnlite-wails/backend/config"
	"huhnlite-wails/backend/db/repo"
	"huhnlite-wails/backend/db/repo_mysql"
	"huhnlite-wails/backend/db/repo_postgres"
)

type DB struct {
	SQL           *sql.DB
	Repo          repo.Querier
	RepoMySQL     *repo_mysql.Queries
	RepoPostgres  *repo_postgres.Queries
	Config        config.Config
	Engine        string
	ActiveConnStr string
	IsTestMode    bool
}
func Connect(cfg config.Config) (*DB, error) {
	var conn *sql.DB
	var err error

	connStr := cfg.DBConnectionString
	if cfg.Test == 1 && cfg.DBConnectionTest != "" {
		connStr = cfg.DBConnectionTest
	}

	if cfg.DBEngine == "sqlite" {
		conn, err = sql.Open("sqlite", connStr)
	} else if cfg.DBEngine == "mysql" {
		// Expecting MariaDB/MySQL DSN
		conn, err = sql.Open("mysql", connStr)
	} else if cfg.DBEngine == "postgres" {
		// Expecting PostgreSQL DSN or URI
		conn, err = sql.Open("postgres", connStr)
	} else {
		return nil, fmt.Errorf("unsupported database engine: %s", cfg.DBEngine)
	}

	if err == nil {
		// Initialize schema if needed / Migrations (Shared for SQLite and MySQL)
		// Renaming SYSTEM to SYSTEM_KZ to avoid reserved word conflicts
		tables := []string{"TEXTE", "TEXT_TYPEN"}
		for _, table := range tables {
			var hasOldColumn bool = hasTableColumn(conn, cfg.DBEngine, table, "SYSTEM")

			if hasOldColumn {
				hasNewColumn := hasTableColumn(conn, cfg.DBEngine, table, "SYSTEM_KZ")

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
				if !hasTableColumn(conn, cfg.DBEngine, table, "SYSTEM_KZ") {
					if _, err := conn.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN SYSTEM_KZ INTEGER NOT NULL DEFAULT 0", table)); err == nil {
						log.Printf("[DB] Successfully added SYSTEM_KZ column to %s (%s)", table, cfg.DBEngine)
					}
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
				{"BEMERKUNG", "TEXT"},
			}
			for _, col := range columns {
				var hasColumn bool = hasTableColumn(conn, cfg.DBEngine, table, col.name)
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
				var hasColumn bool = hasTableColumn(conn, cfg.DBEngine, table, col.name)
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

		// Fix für HERDEN (ZEITSTEMPEL hinzufügen falls fehlend)
		{
			table := "HERDEN"
			columns := []struct {
				name string
				spec string
			}{
				{"ZEITSTEMPEL", "VARCHAR(50) DEFAULT ''"},
			}
			for _, col := range columns {
				var hasColumn bool = hasTableColumn(conn, cfg.DBEngine, table, col.name)
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
				var hasColumn bool = hasTableColumn(conn, cfg.DBEngine, table, col.name)
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
				var hasColumn bool = hasTableColumn(conn, cfg.DBEngine, table, col.name)
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

		// Fix für EILAGER (KLASSE6 und KLASSE7 hinzufügen falls fehlend)
		{
			table := "EILAGER"
			columns := []struct {
				name string
				spec string
			}{
				{"KLASSE6", "VARCHAR(20) DEFAULT ''"},
				{"KLASSE7", "VARCHAR(20) DEFAULT ''"},
			}
			for _, col := range columns {
				var hasColumn bool = hasTableColumn(conn, cfg.DBEngine, table, col.name)
				if !hasColumn {
					log.Printf("[DB] Adding missing column %s to %s (%s)...", col.name, table, cfg.DBEngine)
					if _, err := conn.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, col.name, col.spec)); err == nil {
						log.Printf("[DB] Successfully added %s to %s", col.name, table)
					} else {
						log.Printf("[DB] Error adding %s to %s: %v", col.name, table, err)
					}
				}
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
		SQL:           conn,
		RepoMySQL:     repo_mysql.New(conn),
		RepoPostgres:  repo_postgres.New(conn),
		Config:        cfg,
		Engine:        cfg.DBEngine,
		ActiveConnStr: connStr,
		IsTestMode:    cfg.Test == 1,
	}

	if cfg.DBEngine == "mysql" {
		d.Repo = NewMySQLWrapper(repo.New(conn), d.RepoMySQL, conn)
	} else if cfg.DBEngine == "postgres" {
		d.Repo = NewPostgresWrapper(repo.New(conn), d.RepoPostgres, conn)
	} else {
		d.Repo = repo.New(conn)
	}

	return d, nil
}

func (d *DB) SwitchConnection(connString string, isTest bool) error {
	var conn *sql.DB
	var err error

	if d.Engine == "sqlite" {
		conn, err = sql.Open("sqlite", connString)
	} else if d.Engine == "mysql" {
		conn, err = sql.Open("mysql", connString)
	} else if d.Engine == "postgres" {
		conn, err = sql.Open("postgres", connString)
	} else {
		return fmt.Errorf("unsupported database engine: %s", d.Engine)
	}

	if err != nil {
		return err
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return err
	}

	// Close old connection
	if d.SQL != nil {
		d.SQL.Close()
	}

	d.SQL = conn
	d.RepoMySQL = repo_mysql.New(conn)
	d.RepoPostgres = repo_postgres.New(conn)
	if d.Engine == "mysql" {
		d.Repo = NewMySQLWrapper(repo.New(conn), d.RepoMySQL, conn)
	} else if d.Engine == "postgres" {
		d.Repo = NewPostgresWrapper(repo.New(conn), d.RepoPostgres, conn)
	} else {
		d.Repo = repo.New(conn)
	}

	d.ActiveConnStr = connString
	d.IsTestMode = isTest

	log.Printf("Successfully switched database (TestMode: %v) to: %s", isTest, connString)
	return nil
}

func hasTableColumn(conn *sql.DB, engine, table, colName string) bool {
	if engine == "sqlite" {
		rows, err := conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
		if err != nil {
			return false
		}
		defer rows.Close()
		for rows.Next() {
			var cid int
			var name, dtype string
			var notnull, pk int
			var dflt interface{}
			if err := rows.Scan(&cid, &name, &dtype, &notnull, &dflt, &pk); err == nil {
				if strings.EqualFold(name, colName) {
					return true
				}
			}
		}
		return false
	} else if engine == "postgres" {
		checkQuery := fmt.Sprintf("SELECT column_name FROM information_schema.columns WHERE table_name = '%s' AND column_name = '%s'", strings.ToLower(table), strings.ToLower(colName))
		var dummy string
		err := conn.QueryRow(checkQuery).Scan(&dummy)
		return err == nil
	} else {
		checkQuery := fmt.Sprintf("SHOW COLUMNS FROM %s LIKE '%s'", table, colName)
		rows, err := conn.Query(checkQuery)
		if err != nil {
			return false
		}
		defer rows.Close()
		return rows.Next()
	}
}
