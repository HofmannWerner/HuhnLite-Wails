package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"

	"huhnlite-wails/backend/config"
	"huhnlite-wails/backend/db/repo"
)

type DB struct {
	SQL    *sql.DB
	Repo   repo.Querier
	Config config.Config
}

func Connect(cfg config.Config) (*DB, error) {
	var conn *sql.DB
	var err error

	if cfg.DBEngine == "sqlite" {
		conn, err = sql.Open("sqlite", cfg.DBConnectionString)
		if err == nil {
			// Initialize schema if needed (in a real app, use migrations)
			// For standalone, maybe run schema_sqlite.sql on new file
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

	return &DB{
		SQL:    conn,
		Repo:   repo.New(conn),
		Config: cfg,
	}, nil
}
