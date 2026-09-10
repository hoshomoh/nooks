package profile

import (
	"io"
	"testing"
)

// noEnv is an empty environment, for cases that should fall back to defaults.
func noEnv(string) string { return "" }

// envFrom returns a lookup backed by a map, so a test states its environment inline.
func envFrom(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

func TestParseDefaults(t *testing.T) {
	cfg, err := Parse(nil, noEnv, io.Discard)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if cfg.Addr != DefaultAddr {
		t.Errorf("Addr = %q, want %q", cfg.Addr, DefaultAddr)
	}
	if cfg.Driver != DriverSQLite {
		t.Errorf("Driver = %q, want %q", cfg.Driver, DriverSQLite)
	}
	if cfg.Mode != ModeProd {
		t.Errorf("Mode = %q, want %q", cfg.Mode, ModeProd)
	}
	if got, want := cfg.SQLitePath(), "data/nook.db"; got != want {
		t.Errorf("SQLitePath() = %q, want %q", got, want)
	}
}

func TestParseFlagsBeatEnvironment(t *testing.T) {
	env := envFrom(map[string]string{"NOOK_ADDR": ":9000", "NOOK_DATA": "/from/env"})

	cfg, err := Parse([]string{"--addr", ":7000"}, env, io.Discard)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if cfg.Addr != ":7000" {
		t.Errorf("Addr = %q, want the flag value :7000", cfg.Addr)
	}
	if cfg.Data != "/from/env" {
		t.Errorf("Data = %q, want the environment value when no flag is given", cfg.Data)
	}
}

func TestParseRejectsBadConfig(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"unknown driver", []string{"--driver", "mysql"}},
		{"postgres without dsn", []string{"--driver", "postgres"}},
		{"sqlite with dsn", []string{"--dsn", "postgres://localhost/nook"}},
		{"unknown mode", []string{"--mode", "staging"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Parse(tc.args, noEnv, io.Discard); err == nil {
				t.Fatal("Parse succeeded, want an error")
			}
		})
	}
}

func TestParseAcceptsPostgres(t *testing.T) {
	cfg, err := Parse([]string{"--driver", "postgres", "--dsn", "postgres://localhost/nook"}, noEnv, io.Discard)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if cfg.Driver != DriverPostgres {
		t.Errorf("Driver = %q, want %q", cfg.Driver, DriverPostgres)
	}
}

func TestParseRequiresEnvLookup(t *testing.T) {
	if _, err := Parse(nil, nil, io.Discard); err == nil {
		t.Fatal("Parse with a nil env succeeded, want an error")
	}
}
