package telemetry

import (
	"strings"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

// ConfigureSentry wires the logger package to use Sentry settings provided via application config.
// If no DSN is present, it quietly skips configuration so environment variables can still be used.
func ConfigureSentry(cfg *config.Config) {
	if cfg == nil {
		return
	}
	dsn := strings.TrimSpace(cfg.Sentry.DSN)
	if dsn == "" {
		return
	}

	logger.ConfigureSentry(logger.SentryOptions{
		DSN:              dsn,
		Environment:      coalesce(cfg.Sentry.Environment, cfg.Environment),
		Release:          cfg.Sentry.Release,
		Debug:            cfg.Sentry.Debug,
		EnableTracing:    cfg.Sentry.EnableTracing,
		TracesSampleRate: cfg.Sentry.TracesSampleRate,
	})
}

func coalesce(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
