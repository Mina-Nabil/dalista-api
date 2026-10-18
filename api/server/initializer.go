package server

import (
	"fmt"
	"time"

	"dalista/dalista-be/api/server/logger"
	"dalista/dalista-be/config"

	"github.com/getsentry/sentry-go"
)

// Initialize wires up cross-cutting concerns (Sentry-backed logging) before the
// HTTP server starts. Called once from main().
func Initialize() error {
	cfg := config.GetConfig()

	if err := logger.Init(logger.InitOptions{
		DSN:         cfg.SentryDSN,
		Environment: cfg.AppEnv,
	}); err != nil {
		// Sentry isn't up, so we can't route this through the logger.
		fmt.Printf("Sentry initialization failed: %v\n", err)
		return err
	}

	logger.InfoNoCtx("Initializing Dalista API")

	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)

	return nil
}
