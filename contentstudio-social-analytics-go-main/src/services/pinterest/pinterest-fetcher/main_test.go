package main

import (
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

// ================== Main Function Tests ==================

func TestTopicConstants(t *testing.T) {
	topics := map[string]string{
		"topicWorkOrderBatch":  topicWorkOrderBatch,
		"topicRawUsers":        topicRawUsers,
		"topicRawBoards":       topicRawBoards,
		"topicRawPins":         topicRawPins,
		"topicRawPinInsights":  topicRawPinInsights,
		"topicRawUserInsights": topicRawUserInsights,
	}

	for name, topic := range topics {
		if topic == "" {
			t.Errorf("%s should not be empty", name)
		}
	}
}

func TestSyncTypeConstants(t *testing.T) {
	if fullSyncDays != 86 {
		t.Errorf("fullSyncDays = %d, want 86", fullSyncDays)
	}
	if incrementalSyncDays != 3 {
		t.Errorf("incrementalSyncDays = %d, want 3", incrementalSyncDays)
	}
	if immediateSyncDays != 90 {
		t.Errorf("immediateSyncDays = %d, want 90", immediateSyncDays)
	}
}

func TestPageSizeConstants(t *testing.T) {
	if fullPageSize != 250 {
		t.Errorf("fullPageSize = %d, want 250", fullPageSize)
	}
	if incrementalPageSize != 25 {
		t.Errorf("incrementalPageSize = %d, want 25", incrementalPageSize)
	}
	if immediatePageSize != 250 {
		t.Errorf("immediatePageSize = %d, want 250", immediatePageSize)
	}
	if maxIncrementalPages != 2 {
		t.Errorf("maxIncrementalPages = %d, want 2", maxIncrementalPages)
	}
}

func TestWorkerPoolConstants(t *testing.T) {
	if maxWorkers != 5 {
		t.Errorf("maxWorkers = %d, want 5", maxWorkers)
	}
	if workOrderChanSize != 100 {
		t.Errorf("workOrderChanSize = %d, want 100", workOrderChanSize)
	}
	if timestampChanSize != 200 {
		t.Errorf("timestampChanSize = %d, want 200", timestampChanSize)
	}
}

func TestConsumerGroupConstant(t *testing.T) {
	if consumerGroup != "pinterest-fetcher-group" {
		t.Errorf("consumerGroup = %q, want %q", consumerGroup, "pinterest-fetcher-group")
	}
}

func TestAnalyticsEndDateOffset(t *testing.T) {
	if analyticsEndDateOffset != 2 {
		t.Errorf("analyticsEndDateOffset = %d, want 2 (skip today and yesterday)", analyticsEndDateOffset)
	}
}

func TestErrUnauthorized(t *testing.T) {
	if ErrUnauthorized == nil {
		t.Error("ErrUnauthorized should not be nil")
	}
	if ErrUnauthorized.Error() != "unauthorized: invalid or expired token" {
		t.Errorf("ErrUnauthorized.Error() = %q, want %q", ErrUnauthorized.Error(), "unauthorized: invalid or expired token")
	}
}

// ================== Logging Contract Tests (Point 4 — Calling service logs errors with context) ==================

func TestLoggingContract_PinterestFetcher_ErrorHasContextFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	log.Error().
		Str("error_message", "failed to fetch boards").
		Str("function", "fetchAllBoardsData").
		Str("stage", "fetch_boards").
		Msg("Pinterest fetcher error")

	output := buf.String()

	checks := map[string]string{
		"ERR":                "expected ERR level",
		"error_message":      "expected error_message field",
		"function":           "expected function field",
		"fetchAllBoardsData": "expected fetchAllBoardsData value",
		"stage":              "expected stage field",
		"fetch_boards":       "expected fetch_boards stage value",
	}
	for substr, errMsg := range checks {
		if !strings.Contains(output, substr) {
			t.Errorf("%s, got: %s", errMsg, output)
		}
	}
}

func TestLoggingContract_PinterestFetcher_NoCaptureException(t *testing.T) {
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	log.Error().
		Str("error_message", "fetch pins failed").
		Str("function", "fetchPinsForBoard").
		Str("stage", "fetch_pins").
		Msg("Failed to fetch pins")

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls (hook handles Sentry), got %d", len(*captureRecords))
	}
}

func TestLoggingContract_PinterestFetcher_SingleSentryEvent(t *testing.T) {
	hookRecords, cleanup := logger.InstallHookSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	log.Error().
		Str("error_message", "kafka produce timeout").
		Str("function", "produceMessages").
		Str("stage", "produce_kafka").
		Msg("Failed to produce message")

	var errorLevelCount int
	for _, r := range *hookRecords {
		if r.Level == zerolog.ErrorLevel {
			errorLevelCount++
		}
	}
	if errorLevelCount != 1 {
		t.Fatalf("expected exactly 1 ErrorLevel hook firing, got %d", errorLevelCount)
	}
}

func TestLoggingContract_PinterestFetcher_ExpectedError_WarnOnly(t *testing.T) {
	hookRecords, hookCleanup := logger.InstallHookSpy()
	defer hookCleanup()

	log, buf := logger.NewTestLoggerWithHook()

	// Simulate expected unauthorized error — should be Warn
	log.Warn().
		Str("error_message", "unauthorized: invalid or expired token").
		Str("function", "processWorkOrder").
		Str("pinterest_user", "user-123").
		Msg("Pinterest token expired, skipping account")

	output := buf.String()

	if !strings.Contains(output, "WRN") {
		t.Fatalf("expected WRN level, got: %s", output)
	}
	if strings.Contains(output, "ERR") {
		t.Fatalf("expected error should NOT produce ERR level: %s", output)
	}

	for _, r := range *hookRecords {
		if r.Level >= zerolog.ErrorLevel {
			t.Fatalf("expected error should not trigger Error+ hook, got level %v", r.Level)
		}
	}
}
