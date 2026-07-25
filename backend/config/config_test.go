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
