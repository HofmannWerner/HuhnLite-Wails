package db

import (
	"database/sql"
	"fmt"
	"log"

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
		if err == nil {
			// Initialize schema if needed / Migrations
			if _, err := conn.Exec("ALTER TABLE TEXTE ADD COLUMN SYSTEM INTEGER NOT NULL DEFAULT 0"); err != nil {
				log.Printf("[DB] Note: Could not add SYSTEM to TEXTE (might already exist): %v", err)
			} else {
				log.Println("[DB] Successfully added SYSTEM column to TEXTE (SQLite)")
			}

			if _, err := conn.Exec("ALTER TABLE TEXT_TYPEN ADD COLUMN SYSTEM INTEGER NOT NULL DEFAULT 0"); err != nil {
				log.Printf("[DB] Note: Could not add SYSTEM to TEXT_TYPEN (might already exist): %v", err)
			} else {
				log.Println("[DB] Successfully added SYSTEM column to TEXT_TYPEN (SQLite)")
			}
		}
	} else if cfg.DBEngine == "mysql" {
		// Expecting MariaDB/MySQL DSN
		conn, err = sql.Open("mysql", cfg.DBConnectionString)
	} else {
		return nil, fmt.Errorf("unsupported database engine: %s", cfg.DBEngine)
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
