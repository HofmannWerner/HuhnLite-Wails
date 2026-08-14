package config

import (
	"testing"
)

func TestParseCLIArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantPort    int
		wantMandant int
	}{
		{
			name:        "Standard flag format space separated",
			args:        []string{"-Port", "9001", "--Mandant", "1"},
			wantPort:    9001,
			wantMandant: 1,
		},
		{
			name:        "Single string with spaces as passed by some launchers",
			args:        []string{"-Port 9001 --Mandant 1"},
			wantPort:    9001,
			wantMandant: 1,
		},
		{
			name:        "Combined argument elements with internal spaces",
			args:        []string{"-Port 9001", "--Mandant 1"},
			wantPort:    9001,
			wantMandant: 1,
		},
		{
			name:        "Equals separated flags",
			args:        []string{"-Port=9001", "--Mandant=1"},
			wantPort:    9001,
			wantMandant: 1,
		},
		{
			name:        "Colon separated flags",
			args:        []string{"-Port:9001", "-Mandant:1"},
			wantPort:    9001,
			wantMandant: 1,
		},
		{
			name:        "Positional numeric arguments",
			args:        []string{"9001", "1"},
			wantPort:    9001,
			wantMandant: 1,
		},
		{
			name:        "macOS Finder argument psn mixed with flags",
			args:        []string{"-psn_0_123456", "-Port", "9001", "--Mandant", "1"},
			wantPort:    9001,
			wantMandant: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ov := parseCLIArgs(tt.args)

			if ov.Port == nil {
				t.Errorf("%s: expected Port %d, got nil", tt.name, tt.wantPort)
			} else if *ov.Port != tt.wantPort {
				t.Errorf("%s: expected Port %d, got %d", tt.name, tt.wantPort, *ov.Port)
			}

			if ov.Mandant == nil {
				t.Errorf("%s: expected Mandant %d, got nil", tt.name, tt.wantMandant)
			} else if *ov.Mandant != tt.wantMandant {
				t.Errorf("%s: expected Mandant %d, got %d", tt.name, tt.wantMandant, *ov.Mandant)
			}
		})
	}
}

func TestApplyMandantToDBConnection(t *testing.T) {
	tests := []struct {
		name      string
		engine    string
		connStr   string
		mandantID int
		want      string
	}{
		{
			name:      "MySQL DSN with params",
			engine:    "mysql",
			connStr:   "root:studio@tcp(192.168.178.60:3307)/huhnlite?parseTime=true&allowNativePasswords=true",
			mandantID: 1,
			want:      "root:studio@tcp(192.168.178.60:3307)/huhnlite-1?parseTime=true&allowNativePasswords=true",
		},
		{
			name:      "MySQL DSN test db with params",
			engine:    "mysql",
			connStr:   "root:studio@tcp(192.168.178.60:3307)/huhnlite_test?parseTime=true&allowNativePasswords=true",
			mandantID: 2,
			want:      "root:studio@tcp(192.168.178.60:3307)/huhnlite_test-2?parseTime=true&allowNativePasswords=true",
		},
		{
			name:      "MySQL DSN without params",
			engine:    "mysql",
			connStr:   "root:pass@tcp(localhost:3306)/mydb",
			mandantID: 3,
			want:      "root:pass@tcp(localhost:3306)/mydb-3",
		},
		{
			name:      "PostgreSQL URL prod",
			engine:    "postgres",
			connStr:   "postgres://postgres:post@192.168.178.28:5432/huhnlite-prod?sslmode=disable",
			mandantID: 1,
			want:      "postgres://postgres:post@192.168.178.28:5432/huhnlite-prod-1?sslmode=disable",
		},
		{
			name:      "PostgreSQL URL test mandant 2",
			engine:    "postgres",
			connStr:   "postgres://postgres:post@192.168.178.28:5432/huhnlite-test?sslmode=disable",
			mandantID: 2,
			want:      "postgres://postgres:post@192.168.178.28:5432/huhnlite-test-2?sslmode=disable",
		},
		{
			name:      "PostgreSQL URL without query",
			engine:    "postgres",
			connStr:   "postgresql://user:pass@host:5432/huhnlite",
			mandantID: 1,
			want:      "postgresql://user:pass@host:5432/huhnlite-1",
		},
		{
			name:      "PostgreSQL Key-Value DSN",
			engine:    "postgres",
			connStr:   "host=127.0.0.1 port=5432 user=pg password=pg dbname=huhnlite-prod sslmode=disable",
			mandantID: 1,
			want:      "host=127.0.0.1 port=5432 user=pg password=pg dbname=huhnlite-prod-1 sslmode=disable",
		},
		{
			name:      "SQLite untouched",
			engine:    "sqlite",
			connStr:   "HuhnLite_prod.db",
			mandantID: 1,
			want:      "HuhnLite_prod.db",
		},
		{
			name:      "Mandant ID 0 untouched",
			engine:    "mysql",
			connStr:   "root:studio@tcp(127.0.0.1:3307)/huhnlite",
			mandantID: 0,
			want:      "root:studio@tcp(127.0.0.1:3307)/huhnlite",
		},
		{
			name:      "Empty string untouched",
			engine:    "postgres",
			connStr:   "",
			mandantID: 1,
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyMandantToDBConnection(tt.engine, tt.connStr, tt.mandantID)
			if got != tt.want {
				t.Errorf("ApplyMandantToDBConnection() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMandantConfigSimulation(t *testing.T) {
	// MariaDB simulation
	mariaProd := "root:studio@tcp(192.168.178.60:3307)/huhnlite?parseTime=true&allowNativePasswords=true"
	mariaTest := "root:studio@tcp(192.168.178.60:3307)/huhnlite_test?parseTime=true&allowNativePasswords=true"

	mandant1Prod := ApplyMandantToDBConnection("mysql", mariaProd, 1)
	mandant1Test := ApplyMandantToDBConnection("mysql", mariaTest, 1)
	if mandant1Prod != "root:studio@tcp(192.168.178.60:3307)/huhnlite-1?parseTime=true&allowNativePasswords=true" {
		t.Errorf("unexpected mandant 1 prod DB: %s", mandant1Prod)
	}
	if mandant1Test != "root:studio@tcp(192.168.178.60:3307)/huhnlite_test-1?parseTime=true&allowNativePasswords=true" {
		t.Errorf("unexpected mandant 1 test DB: %s", mandant1Test)
	}

	mandant2Prod := ApplyMandantToDBConnection("mysql", mariaProd, 2)
	mandant2Test := ApplyMandantToDBConnection("mysql", mariaTest, 2)
	if mandant2Prod != "root:studio@tcp(192.168.178.60:3307)/huhnlite-2?parseTime=true&allowNativePasswords=true" {
		t.Errorf("unexpected mandant 2 prod DB: %s", mandant2Prod)
	}
	if mandant2Test != "root:studio@tcp(192.168.178.60:3307)/huhnlite_test-2?parseTime=true&allowNativePasswords=true" {
		t.Errorf("unexpected mandant 2 test DB: %s", mandant2Test)
	}

	// Postgres simulation
	pgProd := "postgres://postgres:post@192.168.178.28:5432/huhnlite-prod?sslmode=disable"
	pgTest := "postgres://postgres:post@192.168.178.28:5432/huhnlite-test?sslmode=disable"

	pgMandant1Prod := ApplyMandantToDBConnection("postgres", pgProd, 1)
	pgMandant1Test := ApplyMandantToDBConnection("postgres", pgTest, 1)
	if pgMandant1Prod != "postgres://postgres:post@192.168.178.28:5432/huhnlite-prod-1?sslmode=disable" {
		t.Errorf("unexpected pg mandant 1 prod DB: %s", pgMandant1Prod)
	}
	if pgMandant1Test != "postgres://postgres:post@192.168.178.28:5432/huhnlite-test-1?sslmode=disable" {
		t.Errorf("unexpected pg mandant 1 test DB: %s", pgMandant1Test)
	}

	pgMandant2Prod := ApplyMandantToDBConnection("postgres", pgProd, 2)
	pgMandant2Test := ApplyMandantToDBConnection("postgres", pgTest, 2)
	if pgMandant2Prod != "postgres://postgres:post@192.168.178.28:5432/huhnlite-prod-2?sslmode=disable" {
		t.Errorf("unexpected pg mandant 2 prod DB: %s", pgMandant2Prod)
	}
	if pgMandant2Test != "postgres://postgres:post@192.168.178.28:5432/huhnlite-test-2?sslmode=disable" {
		t.Errorf("unexpected pg mandant 2 test DB: %s", pgMandant2Test)
	}
}

func TestExchangePath(t *testing.T) {
	cfg := Config{
		ExchangePath: "g:/Meine Ablage/Huhnlite",
	}
	if cfg.ExchangePath != "g:/Meine Ablage/Huhnlite" {
		t.Errorf("expected ExchangePath to be 'g:/Meine Ablage/Huhnlite', got %q", cfg.ExchangePath)
	}
}


