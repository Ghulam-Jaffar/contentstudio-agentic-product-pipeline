package main

import (
	"strings"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/rs/zerolog"
)

// ================== WorkOrderMessage Tests ==================

func TestWorkOrderMessage(t *testing.T) {
	msg := WorkOrderMessage{
		Key:   []byte("test-key"),
		Value: []byte("test-value"),
	}

	if string(msg.Key) != "test-key" {
		t.Fatalf("expected key 'test-key', got '%s'", string(msg.Key))
	}
	if string(msg.Value) != "test-value" {
		t.Fatalf("expected value 'test-value', got '%s'", string(msg.Value))
	}
}

// ================== FetcherMetrics Tests ==================

func TestFetcherMetrics_Initialization(t *testing.T) {
	metrics := &FetcherMetrics{}

	if metrics.BatchesReceived != 0 {
		t.Errorf("BatchesReceived = %d, want 0", metrics.BatchesReceived)
	}
	if metrics.AccountsProcessed != 0 {
		t.Errorf("AccountsProcessed = %d, want 0", metrics.AccountsProcessed)
	}
	if metrics.TweetsProduced != 0 {
		t.Errorf("TweetsProduced = %d, want 0", metrics.TweetsProduced)
	}
	if metrics.InsightsProduced != 0 {
		t.Errorf("InsightsProduced = %d, want 0", metrics.InsightsProduced)
	}
}

// ================== Configuration Tests ==================

func TestFetcherConfig_Values(t *testing.T) {
	cfg := DefaultFetcherConfig()

	tests := []struct {
		name  string
		value interface{}
		want  interface{}
	}{
		{"MaxWorkers", cfg.MaxWorkers, 10},
		{"WorkQueueSize", cfg.WorkQueueSize, 200},
		{"TweetsPerPage", cfg.TweetsPerPage, 100},
		{"IdleTimeout", cfg.IdleTimeout, 15 * time.Minute},
		{"IdleCheckPeriod", cfg.IdleCheckPeriod, 30 * time.Second},
		{"RawPostsTopic", cfg.RawPostsTopic, "raw-twitter-posts"},
		{"RawInsightsTopic", cfg.RawInsightsTopic, "raw-twitter-insights"},
		{"AccountSemaphoreCapacity", cfg.AccountSemaphoreCapacity, int64(1)},
	}

	for _, test := range tests {
		if test.value != test.want {
			t.Errorf("%s = %v, want %v", test.name, test.value, test.want)
		}
	}
}

// ================== Interface Compliance Tests ==================

func TestMockConsumer_Implements_KafkaConsumer(t *testing.T) {
	var _ kafka.Consumer = (*kafka.MockConsumer)(nil)
}

func TestMockProducer_Implements_KafkaProducer(t *testing.T) {
	var _ kafka.Producer = (*kafka.MockProducer)(nil)
}

// ================== Logging Contract Tests ==================

func TestLoggingContract_TwitterFetcher_NoCaptureException(t *testing.T) {
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	// Log an error the way the Twitter fetcher does — Error level only, no CaptureException
	log.Error().
		Str("error_message", "Twitter API rate limit exceeded").
		Str("function", "HandleWorkOrder").
		Str("stage", "fetch_tweets").
		Msg("Failed to fetch tweets")

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls (hook handles Sentry), got %d", len(*captureRecords))
	}
}

func TestLoggingContract_TwitterFetcher_ExpectedError_WarnLevel(t *testing.T) {
	hookRecords, hookCleanup := logger.InstallHookSpy()
	defer hookCleanup()

	log, buf := logger.NewTestLoggerWithHook()

	// Simulate an expected/auth error — should be Warn level
	log.Warn().
		Str("error_message", "invalid or expired token").
		Str("function", "HandleWorkOrder").
		Str("twitter_id", "test-id").
		Msg("Twitter auth error, skipping account")

	output := buf.String()

	if !strings.Contains(output, "WRN") {
		t.Fatalf("expected WRN level in output, got: %s", output)
	}
	if strings.Contains(output, "ERR") {
		t.Fatalf("expected error should NOT produce ERR level: %s", output)
	}

	// Verify no Error-level hook firings
	for _, r := range *hookRecords {
		if r.Level >= zerolog.ErrorLevel {
			t.Fatalf("expected error should not trigger Error+ hook, got level %v", r.Level)
		}
	}
}
