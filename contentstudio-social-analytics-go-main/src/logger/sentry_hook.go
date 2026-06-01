package logger

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	sentry "github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
)

// SentryOptions represent runtime configuration for Sentry initialization.
type SentryOptions struct {
	DSN              string
	Environment      string
	Release          string
	Debug            bool
	EnableTracing    bool
	TracesSampleRate float64
}

var (
	sentryInitOnce          sync.Once
	sentryHookActive        atomic.Bool
	configuredSentryOptions atomic.Pointer[SentryOptions]
)

// ConfigureSentry allows callers to set runtime options sourced from app config before any logger is constructed.
func ConfigureSentry(opts SentryOptions) {
	configuredSentryOptions.Store(&opts)
}

// enableSentryHookFromEnv bootstraps Sentry once the first time a logger is created.
// It prefers explicit runtime configuration (ConfigureSentry) and falls back to environment variables.
func enableSentryHookFromEnv() {
	sentryInitOnce.Do(func() {
		opts, ok := resolveSentryOptions()
		if !ok {
			return
		}

		clientOpts := sentry.ClientOptions{
			Dsn:              opts.DSN,
			AttachStacktrace: true,
			Environment:      opts.Environment,
			Release:          opts.Release,
			Debug:            opts.Debug,
			EnableTracing:    opts.EnableTracing,
		}

		if opts.TracesSampleRate > 0 {
			clientOpts.TracesSampleRate = opts.TracesSampleRate
		}

		if err := sentry.Init(clientOpts); err != nil {
			fmt.Fprintf(os.Stderr, "sentry initialization failed: %v\n", err)
			return
		}

		sentryHookActive.Store(true)
	})
}

func resolveSentryOptions() (SentryOptions, bool) {
	if cfg := configuredSentryOptions.Load(); cfg != nil && strings.TrimSpace(cfg.DSN) != "" {
		return *cfg, true
	}
	return gatherOptionsFromEnv()
}

func gatherOptionsFromEnv() (SentryOptions, bool) {
	dsn := firstNonEmpty(os.Getenv("APP_SENTRY_DSN"), os.Getenv("SENTRY_DSN"))
	if dsn == "" {
		return SentryOptions{}, false
	}

	opts := SentryOptions{
		DSN:         dsn,
		Environment: firstNonEmpty(os.Getenv("APP_SENTRY_ENVIRONMENT"), os.Getenv("SENTRY_ENVIRONMENT"), os.Getenv("APP_ENVIRONMENT")),
		Release:     firstNonEmpty(os.Getenv("APP_SENTRY_RELEASE"), os.Getenv("SENTRY_RELEASE")),
	}

	if traces := firstNonEmpty(os.Getenv("APP_SENTRY_TRACES_SAMPLE_RATE"), os.Getenv("SENTRY_TRACES_SAMPLE_RATE")); traces != "" {
		if rate, err := strconv.ParseFloat(traces, 64); err == nil && rate >= 0 {
			opts.TracesSampleRate = rate
		}
	}
	if enableTracing := firstNonEmpty(os.Getenv("APP_SENTRY_ENABLE_TRACING"), os.Getenv("SENTRY_ENABLE_TRACING")); enableTracing != "" {
		opts.EnableTracing = parseBool(enableTracing)
	}
	if debug := firstNonEmpty(os.Getenv("APP_SENTRY_DEBUG"), os.Getenv("SENTRY_DEBUG")); debug != "" {
		opts.Debug = parseBool(debug)
	}

	return opts, true
}

// isSentryHookEnabled reports whether Sentry successfully initialized.
func isSentryHookEnabled() bool {
	return sentryHookActive.Load()
}

// FlushSentry flushes buffered Sentry events before application shutdown.
func FlushSentry(timeout time.Duration) {
	if !isSentryHookEnabled() {
		return
	}
	sentry.Flush(timeout)
}

// SentryCaptureRecord records a single CaptureException invocation for test assertions.
type SentryCaptureRecord struct {
	Err    error
	Tags   map[string]string
	Extras map[string]interface{}
}

// SentryHookRecord records a single sentry hook firing for test assertions.
type SentryHookRecord struct {
	Level zerolog.Level
	Msg   string
}

// testCaptureSpy, when non-nil, is invoked by CaptureException before the real Sentry call.
// This fires even when Sentry is disabled, so tests can observe behaviour without a DSN.
var testCaptureSpy func(SentryCaptureRecord)

// testHookSpy, when non-nil, is invoked by the sentryHook before the early-return guard.
var testHookSpy func(SentryHookRecord)

// InstallCaptureSpy installs a test spy for CaptureException and returns a pointer to the
// accumulated records plus a cleanup function that MUST be deferred.
func InstallCaptureSpy() (*[]SentryCaptureRecord, func()) {
	records := &[]SentryCaptureRecord{}
	testCaptureSpy = func(r SentryCaptureRecord) { *records = append(*records, r) }
	return records, func() { testCaptureSpy = nil }
}

// InstallHookSpy installs a test spy for the sentryHook and returns a pointer to the
// accumulated records plus a cleanup function.
func InstallHookSpy() (*[]SentryHookRecord, func()) {
	records := &[]SentryHookRecord{}
	testHookSpy = func(r SentryHookRecord) { *records = append(*records, r) }
	return records, func() { testHookSpy = nil }
}

// CaptureException sends a custom exception to Sentry with optional tags and extras.
func CaptureException(err error, tags map[string]string, extras map[string]interface{}) {
	if err == nil {
		return
	}

	// Test spy — fires regardless of Sentry state so tests can observe calls.
	if spy := testCaptureSpy; spy != nil {
		spy(SentryCaptureRecord{Err: err, Tags: tags, Extras: extras})
	}

	if !isSentryHookEnabled() {
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		contextData := make(map[string]interface{}, len(tags)+len(extras))
		for k, v := range tags {
			scope.SetTag(k, v)
			contextData[k] = v
		}
		for k, v := range extras {
			scope.SetExtra(k, v)
			contextData[k] = v
		}
		if len(contextData) > 0 {
			scope.SetContext("logger_context", contextData)
		}
		sentry.CaptureException(err)
	})
}

type sentryHook struct{}

func (s sentryHook) Run(_ *zerolog.Event, level zerolog.Level, msg string) {
	// Test spy — fires before the guard so tests can verify hook activation.
	if spy := testHookSpy; spy != nil {
		spy(SentryHookRecord{Level: level, Msg: msg})
	}

	if !isSentryHookEnabled() || level < zerolog.ErrorLevel {
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("logger", "zerolog")
		scope.SetLevel(convertLevel(level))
		sentry.CaptureMessage(msg)
	})

	if level >= zerolog.FatalLevel {
		FlushSentry(2 * time.Second)
	}
}

func convertLevel(level zerolog.Level) sentry.Level {
	switch level {
	case zerolog.DebugLevel:
		return sentry.LevelDebug
	case zerolog.WarnLevel:
		return sentry.LevelWarning
	case zerolog.ErrorLevel:
		return sentry.LevelError
	case zerolog.FatalLevel, zerolog.PanicLevel:
		return sentry.LevelFatal
	default:
		return sentry.LevelInfo
	}
}

func parseBool(val string) bool {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
