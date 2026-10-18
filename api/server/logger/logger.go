// Package logger is the project's single entry point for application logging.
//
// Every level writes to stdout (so local dev / air output is preserved) and
// emits a structured Sentry log entry at the matching severity. The Error and
// Fatal levels additionally fire sentry.CaptureException so an Issue appears
// in the Sentry UI.
//
// All Sentry-side calls are safe before sentry.Init has run: NewLogger and
// CurrentHub return no-op-friendly objects, which matters for early-boot
// callers like config.loadConfig.
package logger

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

// Field is a typed key/value pair attached to the Sentry log entry (and, for
// Error/Fatal, to the captured exception's tags).
type Field struct {
	Key   string
	Value any
}

// F is a short constructor for a Field, e.g. logger.F("user_id", id).
func F(key string, value any) Field {
	return Field{Key: key, Value: value}
}

func Debug(ctx context.Context, msg string, fields ...Field) {
	emit(ctx, severityDebug, nil, msg, fields)
}

func Info(ctx context.Context, msg string, fields ...Field) {
	emit(ctx, severityInfo, nil, msg, fields)
}

func Warning(ctx context.Context, msg string, fields ...Field) {
	emit(ctx, severityWarning, nil, msg, fields)
}

func Error(ctx context.Context, err error, msg string, fields ...Field) {
	emit(ctx, severityError, err, msg, fields)
}

func DebugNoCtx(msg string, fields ...Field)   { Debug(context.Background(), msg, fields...) }
func InfoNoCtx(msg string, fields ...Field)    { Info(context.Background(), msg, fields...) }
func WarningNoCtx(msg string, fields ...Field) { Warning(context.Background(), msg, fields...) }
func ErrorNoCtx(err error, msg string, fields ...Field) {
	Error(context.Background(), err, msg, fields...)
}

// Fatal logs at error severity, captures the exception, flushes Sentry, then
// terminates the process with exit code 1. Use as a drop-in replacement for
// log.Fatal / log.Fatalf.
func Fatal(err error, msg string, fields ...Field) {
	emit(context.Background(), severityFatal, err, msg, fields)
	sentry.Flush(2 * time.Second)
	os.Exit(1)
}

type severity int

const (
	severityDebug severity = iota
	severityInfo
	severityWarning
	severityError
	severityFatal
)

func (s severity) label() string {
	switch s {
	case severityDebug:
		return "DEBUG"
	case severityInfo:
		return "INFO"
	case severityWarning:
		return "WARN"
	case severityError:
		return "ERROR"
	case severityFatal:
		return "FATAL"
	}
	return "INFO"
}

func emit(ctx context.Context, sev severity, err error, msg string, fields []Field) {
	if ctx == nil {
		ctx = context.Background()
	}

	writeStdout(sev, err, msg, fields)
	emitSentryLog(ctx, sev, msg, fields)

	if sev == severityError || sev == severityFatal {
		captureIssue(ctx, err, msg, fields)
	}
}

func writeStdout(sev severity, err error, msg string, fields []Field) {
	var b strings.Builder
	fmt.Fprintf(&b, "[%s] %s", sev.label(), msg)
	for _, f := range fields {
		fmt.Fprintf(&b, " %s=%v", f.Key, f.Value)
	}
	if err != nil {
		fmt.Fprintf(&b, " err=%v", err)
	}
	log.Println(b.String())
}

func emitSentryLog(ctx context.Context, sev severity, msg string, fields []Field) {
	entry := sentry.NewLogger(ctx)
	var line sentry.LogEntry
	switch sev {
	case severityDebug:
		line = entry.Debug()
	case severityInfo:
		line = entry.Info()
	case severityWarning:
		line = entry.Warn()
	case severityError:
		line = entry.Error()
	case severityFatal:
		line = entry.Fatal()
	default:
		line = entry.Info()
	}
	for _, f := range fields {
		line = applyField(line, f)
	}
	line.Emit(msg)
}

func applyField(line sentry.LogEntry, f Field) sentry.LogEntry {
	switch v := f.Value.(type) {
	case string:
		return line.String(f.Key, v)
	case bool:
		return line.Bool(f.Key, v)
	case int:
		return line.Int(f.Key, v)
	case int64:
		return line.Int64(f.Key, v)
	case float64:
		return line.Float64(f.Key, v)
	case error:
		return line.String(f.Key, v.Error())
	default:
		return line.String(f.Key, fmt.Sprint(v))
	}
}

func captureIssue(ctx context.Context, err error, msg string, fields []Field) {
	if err == nil {
		err = errors.New(msg)
	}
	hub := hubFor(ctx)
	hub.WithScope(func(scope *sentry.Scope) {
		for _, f := range fields {
			scope.SetTag(f.Key, fmt.Sprint(f.Value))
		}
		if msg != "" && msg != err.Error() {
			scope.SetExtra("message", msg)
		}
		hub.CaptureException(err)
	})
}

// hubFor resolves the most specific Sentry hub available: a request-scoped hub
// if a *gin.Context was threaded through ctx (or passed directly), then any hub
// stored on the context, then the global hub.
func hubFor(ctx context.Context) *sentry.Hub {
	if c, ok := ctx.(*gin.Context); ok {
		if h := sentrygin.GetHubFromContext(c); h != nil {
			return h
		}
	}
	if h := sentry.GetHubFromContext(ctx); h != nil {
		return h
	}
	return sentry.CurrentHub()
}
