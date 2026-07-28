package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultUsesEnvironment(t *testing.T) {
	t.Setenv("TINYSCHOOL_API_ADDRESS", "127.0.0.1:9000")
	t.Setenv("TINYSCHOOL_DB_PATH", "/tmp/tinyschool-test.db")

	got := Default()
	if got.Address != "127.0.0.1:9000" {
		t.Fatalf("Address = %q", got.Address)
	}
	if got.DatabasePath != "/tmp/tinyschool-test.db" {
		t.Fatalf("DatabasePath = %q", got.DatabasePath)
	}
	if got.ShutdownTimeout != 10*time.Second {
		t.Fatalf("ShutdownTimeout = %s", got.ShutdownTimeout)
	}
	if got.SessionDuration != 24*time.Hour {
		t.Fatalf("SessionDuration = %s", got.SessionDuration)
	}
}

func TestResolveJWTSecretPersistsGeneratedSecret(t *testing.T) {
	directory := t.TempDir()
	cfg := Config{DatabasePath: filepath.Join(directory, "school.db")}

	first, err := cfg.ResolveJWTSecret()
	if err != nil {
		t.Fatal(err)
	}
	second, err := cfg.ResolveJWTSecret()
	if err != nil {
		t.Fatal(err)
	}
	if first != second || len(first) < 32 {
		t.Fatalf("generated secrets differ or are too short")
	}
	info, err := os.Stat(filepath.Join(directory, ".tinyschool-jwt-secret"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("secret mode = %o", info.Mode().Perm())
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{name: "empty address", config: Config{DatabasePath: "test.db", ShutdownTimeout: time.Second, SessionDuration: time.Hour}},
		{name: "empty database", config: Config{Address: ":8080", ShutdownTimeout: time.Second, SessionDuration: time.Hour}},
		{name: "zero timeout", config: Config{Address: ":8080", DatabasePath: "test.db", SessionDuration: time.Hour}},
		{name: "short secret", config: Config{Address: ":8080", DatabasePath: "test.db", ShutdownTimeout: time.Second, SessionDuration: time.Hour, JWTSecret: "short"}},
		{name: "zero session duration", config: Config{Address: ":8080", DatabasePath: "test.db", ShutdownTimeout: time.Second}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.config.Validate(); err == nil {
				t.Fatal("Validate() error = nil")
			}
		})
	}
}
