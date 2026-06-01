package telemetry

import (
	"testing"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
)

func Test_ConfigureSentry_Table(t *testing.T) {
	cases := []struct {
		name string
		cfg  *config.Config
	}{
		{
			name: "nil config",
			cfg:  nil,
		},
		{
			name: "empty DSN",
			cfg: &config.Config{
				Sentry: config.SentryConfig{
					DSN: "",
				},
			},
		},
		{
			name: "whitespace only DSN",
			cfg: &config.Config{
				Sentry: config.SentryConfig{
					DSN: "   ",
				},
			},
		},
		{
			name: "valid DSN with all options",
			cfg: &config.Config{
				Environment: "test",
				Sentry: config.SentryConfig{
					DSN:              "https://key@sentry.io/123",
					Environment:      "production",
					Release:          "v1.0.0",
					Debug:            true,
					EnableTracing:    true,
					TracesSampleRate: 0.5,
				},
			},
		},
		{
			name: "valid DSN with fallback environment",
			cfg: &config.Config{
				Environment: "staging",
				Sentry: config.SentryConfig{
					DSN:         "https://key@sentry.io/456",
					Environment: "", // should fallback to config.Environment
				},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ConfigureSentry(tc.cfg)
		})
	}
}

func Test_coalesce_Table(t *testing.T) {
	cases := []struct {
		name     string
		values   []string
		expected string
	}{
		{
			name:     "empty values",
			values:   []string{},
			expected: "",
		},
		{
			name:     "all empty strings",
			values:   []string{"", "", ""},
			expected: "",
		},
		{
			name:     "all whitespace strings",
			values:   []string{"   ", "  ", " "},
			expected: "",
		},
		{
			name:     "first non-empty wins",
			values:   []string{"first", "second", "third"},
			expected: "first",
		},
		{
			name:     "skip empty and whitespace",
			values:   []string{"", "  ", "value", "other"},
			expected: "value",
		},
		{
			name:     "single value",
			values:   []string{"only"},
			expected: "only",
		},
		{
			name:     "mixed empty and values",
			values:   []string{"", "valid", ""},
			expected: "valid",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := coalesce(tc.values...)
			if result != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}
