package logger

import (
	"time"

	"github.com/getsentry/sentry-go"
)

// InitOptions carries the values needed to wire Sentry up. Defined here (not
// imported from config) so the config package can call into logger without an
// import cycle.
type InitOptions struct {
	DSN         string
	Environment string
}

// Init initializes the underlying Sentry client. Safe to call from the API
// server bootstrap and from each cmd/* CLI entry point.
func Init(opts InitOptions) error {
	return sentry.Init(sentry.ClientOptions{
		Dsn:         opts.DSN,
		Environment: opts.Environment,
		EnableLogs:  true,
	})
}

// Flush forwards to sentry.Flush so callers don't need to import sentry-go
// directly just to drain the buffer before exit.
func Flush(timeout time.Duration) bool {
	return sentry.Flush(timeout)
}
