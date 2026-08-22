package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"huhnlite-wails/backend/config"
)

func TestMissingSQLiteDBReturnsError(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_sqlite_missing_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "mandant_1", "HuhnLite_prod.db")

	cfg := config.Config{
		DBEngine:           "sqlite",
		DBConnectionString: dbPath,
	}

	database, err := Connect(cfg)
	if err == nil {
		if database != nil && database.SQL != nil {
			database.SQL.Close()
		}
		t.Fatalf("Expected Connect to fail on non-existent sqlite file, but it succeeded")
	}
}

func TestEmptySQLiteDBReturnsError(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_sqlite_empty_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "mandant_1", "HuhnLite_prod.db")
	_ = os.MkdirAll(filepath.Dir(dbPath), 0755)
	f, errCreate := os.Create(dbPath)
	if errCreate != nil {
		t.Fatalf("Failed to create empty file: %v", errCreate)
	}
	f.Close()

	cfg := config.Config{
		DBEngine:           "sqlite",
		DBConnectionString: dbPath,
	}

	database, err := Connect(cfg)
	if err == nil {
		if database != nil && database.SQL != nil {
			database.SQL.Close()
		}
		t.Fatalf("Expected Connect to fail on empty sqlite file without BUCHUNG table, but it succeeded")
	}
}

func TestValidSQLiteDBConnects(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_sqlite_valid_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "mandant_1", "HuhnLite_prod.db")
	_ = os.MkdirAll(filepath.Dir(dbPath), 0755)

	// Create valid table
	rawDB, errOpen := sql.Open("sqlite", dbPath)
	if errOpen != nil {
		t.Fatalf("Failed to open sqlite: %v", errOpen)
	}
	_, errExec := rawDB.Exec("CREATE TABLE BUCHUNG (ID INTEGER PRIMARY KEY, ID_HERDEN INTEGER);")
	rawDB.Close()
	if errExec != nil {
		t.Fatalf("Failed to create test table: %v", errExec)
	}

	cfg := config.Config{
		DBEngine:           "sqlite",
		DBConnectionString: dbPath,
	}

	database, err := Connect(cfg)
	if err != nil {
		t.Fatalf("Expected Connect to succeed on valid sqlite DB: %v", err)
	}
	defer database.SQL.Close()
}
