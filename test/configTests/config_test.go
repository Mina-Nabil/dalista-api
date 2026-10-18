package configTests

import (
	"testing"

	"dalista/dalista-be/config"
)

// TestConfigDefaults is the Phase 0 smoke test: it proves the module builds, the
// test harness runs, and config parsing applies documented defaults. Feature
// phases add real handler/service/resource tests under test/<feature>Tests/.
func TestConfigDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DB_NAME", "dalista_test")
	config.ReloadFromEnv()

	cfg := config.GetConfig()

	if cfg.AppEnv != "development" {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, "development")
	}
	if cfg.DBName != "dalista_test" {
		t.Fatalf("DBName = %q, want %q", cfg.DBName, "dalista_test")
	}
	if cfg.DBPort != "5432" {
		t.Fatalf("DBPort default = %q, want %q", cfg.DBPort, "5432")
	}
	if cfg.RateLimitPerMinute != 40 {
		t.Fatalf("RateLimitPerMinute default = %d, want 40", cfg.RateLimitPerMinute)
	}
	if got := cfg.CorsAllowedOrigins(); len(got) != 1 || got[0] != "http://localhost:3000" {
		t.Fatalf("CorsAllowedOrigins() = %v, want [http://localhost:3000]", got)
	}
	if !cfg.IsStagingEnvironment() {
		t.Fatal("IsStagingEnvironment() = false for development, want true")
	}
}
