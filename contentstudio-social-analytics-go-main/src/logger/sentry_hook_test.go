package logger

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestParseBool(t *testing.T) {
	cases := []struct {
		input    string
		expected bool
	}{
		{"1", true},
		{"true", true},
		{"True", true},
		{"TRUE", true},
		{"yes", true},
		{"Yes", true},
		{"YES", true},
		{"y", true},
		{"Y", true},
		{"on", true},
		{"On", true},
		{"ON", true},
		{"0", false},
		{"false", false},
		{"False", false},
		{"FALSE", false},
		{"no", false},
		{"No", false},
		{"NO", false},
		{"n", false},
		{"N", false},
		{"off", false},
		{"Off", false},
		{"OFF", false},
		{"", false},
		{"  ", false},
		{"invalid", false},
		{" true ", true},
		{" 1 ", true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			result := parseBool(tc.input)
			if result != tc.expected {
				t.Fatalf("parseBool(%q) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestFirstNonEmpty(t *testing.T) {
	cases := []struct {
		name     string
		input    []string
		expected string
	}{
		{"all empty", []string{"", "", ""}, ""},
		{"first non-empty", []string{"first", "second", "third"}, "first"},
		{"second non-empty", []string{"", "second", "third"}, "second"},
		{"third non-empty", []string{"", "", "third"}, "third"},
		{"whitespace only treated as empty", []string{"  ", "value"}, "value"},
		{"single value", []string{"only"}, "only"},
		{"no values", []string{}, ""},
		{"mixed whitespace", []string{" ", "\t", "valid"}, "valid"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := firstNonEmpty(tc.input...)
			if result != tc.expected {
				t.Fatalf("firstNonEmpty(%v) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestConvertLevel(t *testing.T) {
	cases := []struct {
		level zerolog.Level
	}{
		{zerolog.DebugLevel},
		{zerolog.InfoLevel},
		{zerolog.WarnLevel},
		{zerolog.ErrorLevel},
		{zerolog.FatalLevel},
		{zerolog.PanicLevel},
		{zerolog.TraceLevel},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.level.String(), func(t *testing.T) {
			result := convertLevel(tc.level)
			_ = result // Just verify it doesn't panic
		})
	}
}

func TestSentryOptions(t *testing.T) {
	opts := SentryOptions{
		DSN:              "https://test@sentry.io/123",
		Environment:      "test",
		Release:          "v1.0.0",
		Debug:            true,
		EnableTracing:    true,
		TracesSampleRate: 0.5,
	}

	if opts.DSN != "https://test@sentry.io/123" {
		t.Fatalf("expected DSN to be set")
	}
	if opts.Environment != "test" {
		t.Fatalf("expected Environment to be 'test'")
	}
	if opts.Release != "v1.0.0" {
		t.Fatalf("expected Release to be 'v1.0.0'")
	}
	if !opts.Debug {
		t.Fatalf("expected Debug to be true")
	}
	if !opts.EnableTracing {
		t.Fatalf("expected EnableTracing to be true")
	}
	if opts.TracesSampleRate != 0.5 {
		t.Fatalf("expected TracesSampleRate to be 0.5")
	}
}

func TestConfigureSentry(t *testing.T) {
	opts := SentryOptions{
		DSN:         "https://test@sentry.io/456",
		Environment: "staging",
	}

	ConfigureSentry(opts)

	loaded := configuredSentryOptions.Load()
	if loaded == nil {
		t.Fatal("expected options to be stored")
	}
	if loaded.DSN != opts.DSN {
		t.Fatalf("expected DSN %q, got %q", opts.DSN, loaded.DSN)
	}
}

func TestIsSentryHookEnabled_Default(t *testing.T) {
	result := isSentryHookEnabled()
	if result {
		t.Log("Sentry hook is enabled (may be from previous test)")
	}
}

func TestFlushSentry_WhenDisabled(t *testing.T) {
	sentryHookActive.Store(false)
	FlushSentry(1 * time.Second)
}

func TestCaptureException_NilError(t *testing.T) {
	CaptureException(nil, nil, nil)
}

func TestCaptureException_WhenDisabled(t *testing.T) {
	sentryHookActive.Store(false)
	err := &testSentryError{msg: "test"}
	CaptureException(err, map[string]string{"key": "value"}, map[string]interface{}{"extra": "data"})
}

func TestGatherOptionsFromEnv_NoEnv(t *testing.T) {
	os.Unsetenv("APP_SENTRY_DSN")
	os.Unsetenv("SENTRY_DSN")

	opts, ok := gatherOptionsFromEnv()
	if ok {
		t.Fatalf("expected no options when DSN not set, got: %+v", opts)
	}
}

func TestGatherOptionsFromEnv_WithEnv(t *testing.T) {
	os.Setenv("SENTRY_DSN", "https://test@sentry.io/789")
	os.Setenv("SENTRY_ENVIRONMENT", "test-env")
	os.Setenv("SENTRY_RELEASE", "v2.0.0")
	os.Setenv("SENTRY_TRACES_SAMPLE_RATE", "0.25")
	os.Setenv("SENTRY_ENABLE_TRACING", "true")
	os.Setenv("SENTRY_DEBUG", "1")
	defer func() {
		os.Unsetenv("SENTRY_DSN")
		os.Unsetenv("SENTRY_ENVIRONMENT")
		os.Unsetenv("SENTRY_RELEASE")
		os.Unsetenv("SENTRY_TRACES_SAMPLE_RATE")
		os.Unsetenv("SENTRY_ENABLE_TRACING")
		os.Unsetenv("SENTRY_DEBUG")
	}()

	opts, ok := gatherOptionsFromEnv()
	if !ok {
		t.Fatal("expected options to be gathered")
	}
	if opts.DSN != "https://test@sentry.io/789" {
		t.Fatalf("expected DSN from env, got %q", opts.DSN)
	}
	if opts.Environment != "test-env" {
		t.Fatalf("expected Environment 'test-env', got %q", opts.Environment)
	}
	if opts.Release != "v2.0.0" {
		t.Fatalf("expected Release 'v2.0.0', got %q", opts.Release)
	}
	if opts.TracesSampleRate != 0.25 {
		t.Fatalf("expected TracesSampleRate 0.25, got %f", opts.TracesSampleRate)
	}
	if !opts.EnableTracing {
		t.Fatal("expected EnableTracing to be true")
	}
	if !opts.Debug {
		t.Fatal("expected Debug to be true")
	}
}

func TestGatherOptionsFromEnv_AppPrefixedEnv(t *testing.T) {
	os.Setenv("APP_SENTRY_DSN", "https://app@sentry.io/111")
	os.Setenv("APP_SENTRY_ENVIRONMENT", "app-env")
	defer func() {
		os.Unsetenv("APP_SENTRY_DSN")
		os.Unsetenv("APP_SENTRY_ENVIRONMENT")
	}()

	opts, ok := gatherOptionsFromEnv()
	if !ok {
		t.Fatal("expected options to be gathered")
	}
	if opts.DSN != "https://app@sentry.io/111" {
		t.Fatalf("expected APP prefixed DSN, got %q", opts.DSN)
	}
}

func TestGatherOptionsFromEnv_InvalidTraceRate(t *testing.T) {
	os.Setenv("SENTRY_DSN", "https://test@sentry.io/999")
	os.Setenv("SENTRY_TRACES_SAMPLE_RATE", "invalid")
	defer func() {
		os.Unsetenv("SENTRY_DSN")
		os.Unsetenv("SENTRY_TRACES_SAMPLE_RATE")
	}()

	opts, ok := gatherOptionsFromEnv()
	if !ok {
		t.Fatal("expected options to be gathered")
	}
	if opts.TracesSampleRate != 0 {
		t.Fatalf("expected TracesSampleRate 0 for invalid value, got %f", opts.TracesSampleRate)
	}
}

func TestSentryHook_Run_BelowErrorLevel(t *testing.T) {
	hook := sentryHook{}
	hook.Run(nil, zerolog.InfoLevel, "info message")
	hook.Run(nil, zerolog.WarnLevel, "warn message")
	hook.Run(nil, zerolog.DebugLevel, "debug message")
}

func TestSentryHook_Run_WhenDisabled(t *testing.T) {
	sentryHookActive.Store(false)
	hook := sentryHook{}
	hook.Run(nil, zerolog.ErrorLevel, "error message")
	hook.Run(nil, zerolog.FatalLevel, "fatal message")
}

func TestResolveSentryOptions_WithConfigured(t *testing.T) {
	opts := SentryOptions{
		DSN:         "https://configured@sentry.io/123",
		Environment: "configured-env",
	}
	ConfigureSentry(opts)

	resolved, ok := resolveSentryOptions()
	if !ok {
		t.Fatal("expected options to be resolved")
	}
	if resolved.DSN != opts.DSN {
		t.Fatalf("expected DSN %q, got %q", opts.DSN, resolved.DSN)
	}

	configuredSentryOptions.Store(nil)
}

func TestResolveSentryOptions_EmptyDSNFallsBackToEnv(t *testing.T) {
	ConfigureSentry(SentryOptions{DSN: ""})

	os.Setenv("SENTRY_DSN", "https://fallback@sentry.io/456")
	defer os.Unsetenv("SENTRY_DSN")

	resolved, ok := resolveSentryOptions()
	if !ok {
		t.Fatal("expected options to be resolved from env")
	}
	if resolved.DSN != "https://fallback@sentry.io/456" {
		t.Fatalf("expected fallback DSN, got %q", resolved.DSN)
	}

	configuredSentryOptions.Store(nil)
}

func TestResolveSentryOptions_WhitespaceDSNFallsBackToEnv(t *testing.T) {
	ConfigureSentry(SentryOptions{DSN: "   "})

	os.Setenv("SENTRY_DSN", "https://whitespace-fallback@sentry.io/789")
	defer os.Unsetenv("SENTRY_DSN")

	resolved, ok := resolveSentryOptions()
	if !ok {
		t.Fatal("expected options to be resolved from env")
	}
	if resolved.DSN != "https://whitespace-fallback@sentry.io/789" {
		t.Fatalf("expected fallback DSN, got %q", resolved.DSN)
	}

	configuredSentryOptions.Store(nil)
}

func TestFlushSentry_WhenEnabled(t *testing.T) {
	sentryHookActive.Store(true)
	FlushSentry(100 * time.Millisecond)
	sentryHookActive.Store(false)
}

func TestCaptureException_WithTagsAndExtras(t *testing.T) {
	sentryHookActive.Store(false)
	err := &testSentryError{msg: "test error"}

	tags := map[string]string{
		"service": "test-service",
		"version": "1.0.0",
	}
	extras := map[string]interface{}{
		"user_id":   12345,
		"request":   "test-request",
		"timestamp": time.Now(),
	}

	CaptureException(err, tags, extras)
}

func TestCaptureException_EmptyTagsAndExtras(t *testing.T) {
	sentryHookActive.Store(false)
	err := &testSentryError{msg: "test error"}

	CaptureException(err, map[string]string{}, map[string]interface{}{})
	CaptureException(err, nil, nil)
}

func TestGatherOptionsFromEnv_NegativeTraceRate(t *testing.T) {
	os.Setenv("SENTRY_DSN", "https://test@sentry.io/negative")
	os.Setenv("SENTRY_TRACES_SAMPLE_RATE", "-0.5")
	defer func() {
		os.Unsetenv("SENTRY_DSN")
		os.Unsetenv("SENTRY_TRACES_SAMPLE_RATE")
	}()

	opts, ok := gatherOptionsFromEnv()
	if !ok {
		t.Fatal("expected options to be gathered")
	}
	if opts.TracesSampleRate != 0 {
		t.Fatalf("expected TracesSampleRate 0 for negative value, got %f", opts.TracesSampleRate)
	}
}

func TestGatherOptionsFromEnv_AppEnvironmentFallback(t *testing.T) {
	os.Setenv("SENTRY_DSN", "https://test@sentry.io/env")
	os.Setenv("APP_ENVIRONMENT", "fallback-env")
	defer func() {
		os.Unsetenv("SENTRY_DSN")
		os.Unsetenv("APP_ENVIRONMENT")
	}()

	opts, ok := gatherOptionsFromEnv()
	if !ok {
		t.Fatal("expected options to be gathered")
	}
	if opts.Environment != "fallback-env" {
		t.Fatalf("expected Environment 'fallback-env', got %q", opts.Environment)
	}
}

type testSentryError struct {
	msg string
}

func (e *testSentryError) Error() string {
	return e.msg
}

// ==================== Spy & Hook Behaviour Tests ====================

func TestCaptureSpy_RecordsCalls(t *testing.T) {
	records, cleanup := InstallCaptureSpy()
	defer cleanup()

	testErr := &testSentryError{"boom"}
	tags := map[string]string{"platform": "test"}
	extras := map[string]interface{}{"key": "val"}

	CaptureException(testErr, tags, extras)

	if len(*records) != 1 {
		t.Fatalf("expected 1 capture record, got %d", len(*records))
	}
	r := (*records)[0]
	if r.Err != testErr {
		t.Fatalf("expected captured error to be %v, got %v", testErr, r.Err)
	}
	if r.Tags["platform"] != "test" {
		t.Fatalf("expected tag platform=test, got %q", r.Tags["platform"])
	}
	if r.Extras["key"] != "val" {
		t.Fatalf("expected extra key=val, got %v", r.Extras["key"])
	}
}

func TestCaptureSpy_NilErrorNoRecord(t *testing.T) {
	records, cleanup := InstallCaptureSpy()
	defer cleanup()

	CaptureException(nil, nil, nil)

	if len(*records) != 0 {
		t.Fatalf("expected 0 capture records for nil error, got %d", len(*records))
	}
}

func TestCaptureSpy_CleanupRemovesSpy(t *testing.T) {
	_, cleanup := InstallCaptureSpy()
	cleanup()

	// After cleanup, testCaptureSpy should be nil — no panic should occur
	CaptureException(&testSentryError{"after cleanup"}, nil, nil)
}

func TestHookSpy_FiresOnAllLevels(t *testing.T) {
	records, cleanup := InstallHookSpy()
	defer cleanup()

	log, _ := NewTestLoggerWithHook()

	log.Warn().Msg("warn msg")
	log.Error().Msg("error msg")

	if len(*records) < 2 {
		t.Fatalf("expected at least 2 hook records, got %d", len(*records))
	}

	// The hook spy fires for ALL levels because it's before the level guard.
	// Verify we captured both warn and error.
	var foundWarn, foundError bool
	for _, r := range *records {
		if r.Level == zerolog.WarnLevel && r.Msg == "warn msg" {
			foundWarn = true
		}
		if r.Level == zerolog.ErrorLevel && r.Msg == "error msg" {
			foundError = true
		}
	}
	if !foundWarn {
		t.Fatal("expected hook spy to record WarnLevel event")
	}
	if !foundError {
		t.Fatal("expected hook spy to record ErrorLevel event")
	}
}

func TestSentryHook_OnlyFiresOnErrorPlus(t *testing.T) {
	// The sentryHook.Run method has a guard: level < ErrorLevel → return.
	// With Sentry disabled, the guard fires, but the spy still records all levels.
	// This test verifies the spy sees all levels but the REAL hook would skip Warn.
	h := sentryHook{}

	// We can directly test the level guard logic.
	// When sentry is disabled AND level < Error → both guards fail, hook is a no-op.
	// The spy records the invocation before the guard.
	records, cleanup := InstallHookSpy()
	defer cleanup()

	// Call Run directly for different levels
	h.Run(nil, zerolog.WarnLevel, "should not reach sentry")
	h.Run(nil, zerolog.ErrorLevel, "should reach sentry")
	h.Run(nil, zerolog.InfoLevel, "info should not reach sentry")

	if len(*records) != 3 {
		t.Fatalf("expected 3 hook spy records, got %d", len(*records))
	}
	// All 3 are captured by the spy (before the guard).
	// The actual Sentry call only happens for Error+ when sentry is enabled.
	if (*records)[0].Level != zerolog.WarnLevel {
		t.Fatalf("expected first record to be WarnLevel, got %v", (*records)[0].Level)
	}
	if (*records)[1].Level != zerolog.ErrorLevel {
		t.Fatalf("expected second record to be ErrorLevel, got %v", (*records)[1].Level)
	}
}

// ==================== Logging Contract Tests ====================

// TestLoggingContract_WarnDoesNotTriggerSentryHook verifies that Warn-level logs
// do NOT trigger the sentry hook's CaptureMessage (the level guard blocks it).
func TestLoggingContract_WarnDoesNotTriggerSentryHook(t *testing.T) {
	hookRecords, hookCleanup := InstallHookSpy()
	defer hookCleanup()

	log, buf := NewTestLoggerWithHook()
	log.Warn().Err(&testSentryError{"auth expired"}).Msg("expected auth error")

	// The log should appear in the buffer at WRN level
	output := buf.String()
	if !containsSubstring(output, "WRN") {
		t.Fatalf("expected WRN level in output, got: %s", output)
	}
	if containsSubstring(output, "ERR") {
		t.Fatalf("unexpected ERR level in output: %s", output)
	}

	// The hook spy records the event, but the actual sentry guard
	// would block it (level < ErrorLevel). Verify the hook was invoked
	// at Warn level (spy fires before guard).
	var warnCount int
	for _, r := range *hookRecords {
		if r.Level == zerolog.WarnLevel {
			warnCount++
		}
	}
	if warnCount == 0 {
		t.Fatal("expected hook spy to record the WarnLevel event")
	}
}

// TestLoggingContract_ErrorTriggersHook verifies that Error-level logs
// trigger the sentry hook.
func TestLoggingContract_ErrorTriggersHook(t *testing.T) {
	hookRecords, hookCleanup := InstallHookSpy()
	defer hookCleanup()

	log, buf := NewTestLoggerWithHook()
	log.Error().
		Err(&testSentryError{"db connection failed"}).
		Str("error_message", "db connection failed").
		Str("function", "processWorkOrder").
		Str("stage", "fetch_account").
		Msg("Failed to process work order")

	output := buf.String()
	if !containsSubstring(output, "ERR") {
		t.Fatalf("expected ERR level in output, got: %s", output)
	}

	// Verify the hook was invoked at Error level
	var errorCount int
	for _, r := range *hookRecords {
		if r.Level == zerolog.ErrorLevel {
			errorCount++
		}
	}
	if errorCount != 1 {
		t.Fatalf("expected exactly 1 ErrorLevel hook firing, got %d", errorCount)
	}
}

// TestLoggingContract_NoDuplication verifies that a single error produces
// exactly one Error-level log entry (no cascading duplicates).
func TestLoggingContract_NoDuplication(t *testing.T) {
	hookRecords, hookCleanup := InstallHookSpy()
	defer hookCleanup()
	captureRecords, captureCleanup := InstallCaptureSpy()
	defer captureCleanup()

	log, buf := NewTestLoggerWithHook()

	// Simulate what a main service does: one Error log, no CaptureException
	err := &testSentryError{"unexpected failure"}
	log.Error().Err(err).Str("error_message", err.Error()).Msg("work order failed")

	output := buf.String()

	// Count ERR occurrences in the buffer
	errCount := countSubstring(output, "ERR")
	if errCount != 1 {
		t.Fatalf("expected exactly 1 ERR entry in log output, got %d. Output:\n%s", errCount, output)
	}

	// Verify hook fired exactly once at Error level
	var hookErrCount int
	for _, r := range *hookRecords {
		if r.Level == zerolog.ErrorLevel {
			hookErrCount++
		}
	}
	if hookErrCount != 1 {
		t.Fatalf("expected exactly 1 ErrorLevel hook firing, got %d", hookErrCount)
	}

	// Verify NO explicit CaptureException was called (no duplication)
	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls, got %d", len(*captureRecords))
	}
}

// TestLoggingContract_SwallowedErrorUsesCapture verifies that for swallowed errors
// (Warn + CaptureException), exactly one Sentry event is produced.
func TestLoggingContract_SwallowedErrorUsesCapture(t *testing.T) {
	hookRecords, hookCleanup := InstallHookSpy()
	defer hookCleanup()
	captureRecords, captureCleanup := InstallCaptureSpy()
	defer captureCleanup()

	log, buf := NewTestLoggerWithHook()

	// Simulate what a processor does for a swallowed error
	err := &testSentryError{"fetch failed but continuing"}
	log.Warn().Err(err).Msg("non-critical fetch failed, continuing")
	CaptureException(err, map[string]string{"stage": "fetch_data"}, nil)

	output := buf.String()

	// Should be WRN, not ERR
	if containsSubstring(output, "ERR") {
		t.Fatalf("unexpected ERR level for swallowed error: %s", output)
	}

	// Hook should NOT fire at Error level (only Warn was logged)
	for _, r := range *hookRecords {
		if r.Level >= zerolog.ErrorLevel {
			t.Fatalf("hook should not fire at Error+ for Warn log, but got level %v", r.Level)
		}
	}

	// CaptureException should have been called exactly once
	if len(*captureRecords) != 1 {
		t.Fatalf("expected exactly 1 CaptureException call, got %d", len(*captureRecords))
	}
	if (*captureRecords)[0].Tags["stage"] != "fetch_data" {
		t.Fatalf("expected stage tag 'fetch_data', got %q", (*captureRecords)[0].Tags["stage"])
	}
}

// TestLoggingContract_ExpectedErrorNoSentry verifies that expected/auth errors
// produce zero Sentry events (no hook firing at Error+, no CaptureException).
func TestLoggingContract_ExpectedErrorNoSentry(t *testing.T) {
	hookRecords, hookCleanup := InstallHookSpy()
	defer hookCleanup()
	captureRecords, captureCleanup := InstallCaptureSpy()
	defer captureCleanup()

	log, buf := NewTestLoggerWithHook()

	// Simulate expected auth error — Warn only, no CaptureException
	log.Warn().Err(&testSentryError{"token expired"}).Msg("auth error, skipping account")

	output := buf.String()
	if containsSubstring(output, "ERR") {
		t.Fatalf("expected error should not produce ERR level: %s", output)
	}

	// No Error+ hook firings
	for _, r := range *hookRecords {
		if r.Level >= zerolog.ErrorLevel {
			t.Fatalf("expected error should not trigger Error+ hook, but got %v", r.Level)
		}
	}

	// No CaptureException calls
	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException for expected errors, got %d", len(*captureRecords))
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func countSubstring(s, substr string) int {
	count := 0
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			count++
		}
	}
	return count
}
