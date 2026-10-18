package config

import (
	"errors"
	"net/url"
	"os"
	"strings"

	"dalista/dalista-be/api/server/logger"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config is the single source of truth for runtime configuration, parsed from
// environment variables (with .env support for local development). Mirrors the
// Lumina pattern: one Config struct, one cached GetConfig() accessor.
//
// Fields are grouped by concern. Integration credentials (SMS Misr, FCM, Apple
// App Store, Google Play, OAuth) are declared here so later phases can wire them
// without touching bootstrap, but none are `required` yet — the Phase 0 skeleton
// must boot with only DB + core settings present.
type Config struct {
	// --- Core ---
	AppEnv      string `env:"APP_ENV" envDefault:"development"`
	SentryDSN   string `env:"SENTRY_DSN"`
	FrontEndURL string `env:"FRONT_END_URL" envDefault:"http://localhost:3000"`
	// CORSAllowAll reflects any Origin for credentialed CORS (local debugging only).
	CORSAllowAll bool `env:"CORS_ALLOW_ALL"`

	// --- Auth / JWT ---
	JWTKey         string `env:"JWT_KEY"`
	JWTExpiryHours int    `env:"JWT_EXPIRY_HOURS" envDefault:"8"`

	// --- Database ---
	DBHost     string `env:"DB_HOST" envDefault:"localhost"`
	DBPort     string `env:"DB_PORT" envDefault:"5432"`
	DBUser     string `env:"DB_USER" envDefault:"postgres"`
	DBPassword string `env:"DB_PASSWORD"`
	DBName     string `env:"DB_NAME" envDefault:"dalista"`
	DBSSLMode  string `env:"DB_SSLMODE" envDefault:"disable"`

	// --- Cache (Redis) ---
	// Available in local dev via docker-compose; wired into services when a
	// caching/real-time-sync need lands (see CLAUDE.md open items).
	RedisHost     string `env:"REDIS_HOST" envDefault:"localhost"`
	RedisPort     string `env:"REDIS_PORT" envDefault:"6379"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	RedisDB       int    `env:"REDIS_DB" envDefault:"0"`

	// --- Integrations (wired in later phases; declared now, not yet required) ---
	// SMS Misr — OTP delivery + admin-initiated SMS (Phase 1 / Phase 9).
	SMSMisrUsername string `env:"SMS_MISR_USERNAME"`
	SMSMisrPassword string `env:"SMS_MISR_PASSWORD"`
	SMSMisrSender   string `env:"SMS_MISR_SENDER"`
	// Firebase Cloud Messaging — default push channel (Phase 9).
	FCMCredentialsJSON string `env:"FCM_CREDENTIALS_JSON"`
	// Apple App Store Server API — IAP receipt validation + notifications V2 (Phase 3).
	AppleIAPKeyID      string `env:"APPLE_IAP_KEY_ID"`
	AppleIAPIssuerID   string `env:"APPLE_IAP_ISSUER_ID"`
	AppleIAPPrivateKey string `env:"APPLE_IAP_PRIVATE_KEY"`
	AppleIAPBundleID   string `env:"APPLE_IAP_BUNDLE_ID"`
	// AppleIAPEnvironment selects the App Store Server API host: "sandbox" (default) or "production".
	AppleIAPEnvironment string `env:"APPLE_IAP_ENVIRONMENT" envDefault:"sandbox"`
	// AppleRootCAPEM is Apple's Root CA G3 certificate (PEM) used to pin the
	// trust root when verifying Server Notifications V2 JWS signatures. When
	// empty, signature + internal chain checks still run but the root is not
	// pinned — configure this in staging/production (flagged in DAL-164).
	AppleRootCAPEM string `env:"APPLE_ROOT_CA_PEM"`
	// Google Play Developer API — purchase verification + RTDN (Phase 3).
	GooglePlayCredentialsJSON string `env:"GOOGLE_PLAY_CREDENTIALS_JSON"`
	GooglePlayPackageName     string `env:"GOOGLE_PLAY_PACKAGE_NAME"`
	// OAuth — Google + Facebook social sign-in token verification (Phase 1).
	GoogleOAuthClientID string `env:"GOOGLE_OAUTH_CLIENT_ID"`
	FacebookAppID       string `env:"FACEBOOK_APP_ID"`
	FacebookAppSecret   string `env:"FACEBOOK_APP_SECRET"`

	// RateLimitPerMinute caps requests per client IP per minute at the API edge
	// (token bucket, burst = limit). 0 disables the middleware.
	RateLimitPerMinute int `env:"RATE_LIMIT_PER_MINUTE" envDefault:"40"`
}

var cfg *Config

func GetConfig() Config {
	if cfg == nil {
		loadConfig()
	}
	return *cfg
}

// ReloadFromEnv drops cached config so the next GetConfig() re-parses the environment.
func ReloadFromEnv() {
	cfg = nil
	loadConfig()
}

// CorsAllowedOrigins returns a single allowed Origin (scheme://host[:port]) derived from FRONT_END_URL.
func (c *Config) CorsAllowedOrigins() []string {
	return []string{corsOriginFromFrontEndURL(c.FrontEndURL)}
}

// IsStagingEnvironment reports whether the app is running in staging or development.
func (c *Config) IsStagingEnvironment() bool {
	return strings.EqualFold(strings.TrimSpace(c.AppEnv), "staging") ||
		strings.EqualFold(strings.TrimSpace(c.AppEnv), "development")
}

func corsOriginFromFrontEndURL(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "http://localhost:3000"
	}
	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return strings.TrimRight(s, "/")
	}
	return u.Scheme + "://" + u.Host
}

func loadConfig() {
	// Load local .env for developer commands; existing env vars still take precedence.
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		logger.WarningNoCtx("failed to load .env", logger.F("err", err.Error()))
	}

	cfg = &Config{}
	if err := env.Parse(cfg); err != nil {
		logger.Fatal(err, "Failed to load config")
	}
}
