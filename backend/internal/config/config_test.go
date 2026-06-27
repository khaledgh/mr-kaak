package config

import (
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// With a clean-ish env, development defaults should load without error.
	t.Setenv("APP_ENV", "development")
	t.Setenv("DB_NAME", "restaurant")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.HTTP.Port != 8080 {
		t.Errorf("default port = %d, want 8080", cfg.HTTP.Port)
	}
	if cfg.Redis.CacheDB == cfg.Redis.DB {
		t.Errorf("cache DB index should differ from queue DB index")
	}
}

func TestProductionRejectsWeakSecrets(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DB_NAME", "restaurant")
	t.Setenv("JWT_ACCESS_SECRET", "dev-access-secret-change-me")
	t.Setenv("JWT_REFRESH_SECRET", "dev-refresh-secret-change-me")
	t.Setenv("MEILI_MASTER_KEY", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "*")

	if _, err := Load(); err == nil {
		t.Fatal("expected production validation to fail on weak secrets / wildcard CORS")
	}
}

func TestInvalidDurationIsReported(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DB_NAME", "restaurant")
	t.Setenv("HTTP_READ_TIMEOUT", "not-a-duration")

	if _, err := Load(); err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestDSN(t *testing.T) {
	d := DB{User: "u", Password: "p", Host: "h", Port: 3306, Name: "db", Params: "x=1"}
	want := "u:p@tcp(h:3306)/db?x=1"
	if got := d.DSN(); got != want {
		t.Errorf("DSN() = %q, want %q", got, want)
	}
}
