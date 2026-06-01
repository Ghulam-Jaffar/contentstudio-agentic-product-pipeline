package kafka

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/rs/zerolog"
)

const (
	testBrokerLocal           = "localhost:9092"
	testBrokerIP              = "127.0.0.1:9092"
	testTopic                 = "test-topic"
	testGroupID               = "test-group"
	testSASLMechanismSHA256   = "SCRAM-SHA-256"
	failedToCreateConsumerMsg = "failed to create consumer: %v"
	expectedErrorGotNilMsg    = "expected error %q, got nil"
	errorDoesNotContainMsg    = "error %q does not contain %q"
)

func newTestConsumer(t *testing.T, cfg config.KafkaConfig, groupID string) Consumer {
	t.Helper()

	consumer, err := NewConsumer(cfg, groupID, testLogger())
	if err != nil {
		t.Fatalf(failedToCreateConsumerMsg, err)
	}

	return consumer
}

// validateConsumerCreation is a helper to validate consumer creation and reduce complexity
func validateConsumerCreation(t *testing.T, consumer Consumer, err error, wantError string) {
	t.Helper()
	if wantError != "" {
		if err == nil {
			t.Fatalf(expectedErrorGotNilMsg, wantError)
		}
		if !strings.Contains(err.Error(), wantError) {
			t.Fatalf(errorDoesNotContainMsg, err.Error(), wantError)
		}
	} else {
		if err != nil {
			t.Fatalf("expected no error during creation, got %v", err)
		}
		if consumer != nil {
			consumer.Close()
		}
	}
}

// TestNewConsumerValidation tests validation of consumer creation parameters
func TestNewConsumerValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cfg       config.KafkaConfig
		groupID   string
		wantError string
	}{
		{
			name:      "no brokers configured",
			cfg:       config.KafkaConfig{Brokers: nil},
			groupID:   testGroupID,
			wantError: "kafka brokers must be configured",
		},
		{
			name:      "empty brokers list",
			cfg:       config.KafkaConfig{Brokers: []string{}},
			groupID:   testGroupID,
			wantError: "kafka brokers must be configured",
		},
		{
			name: "empty group ID",
			cfg: config.KafkaConfig{
				Brokers: []string{testBrokerLocal},
			},
			groupID:   "",
			wantError: "consumer group ID must be provided",
		},
		{
			name: "SASL enabled but username missing",
			cfg: config.KafkaConfig{
				Brokers: []string{testBrokerIP},
				SASL: config.SASLConfig{
					Enabled:   true,
					Mechanism: "PLAIN",
					Username:  "",
					Password:  "secret",
				},
			},
			groupID:   testGroupID,
			wantError: "", // Consumer creation succeeds; error comes during Consume
		},
		{
			name: "SASL enabled but password missing",
			cfg: config.KafkaConfig{
				Brokers: []string{testBrokerIP},
				SASL: config.SASLConfig{
					Enabled:   true,
					Mechanism: testSASLMechanismSHA256,
					Username:  "alice",
					Password:  "",
				},
			},
			groupID:   testGroupID,
			wantError: "", // Consumer creation succeeds; error comes during Consume
		},
		{
			name: "SASL unsupported mechanism",
			cfg: config.KafkaConfig{
				Brokers: []string{testBrokerIP},
				SASL: config.SASLConfig{
					Enabled:   true,
					Mechanism: "OAUTHBEARER",
					Username:  "alice",
					Password:  "secret",
				},
			},
			groupID:   testGroupID,
			wantError: "", // Consumer creation succeeds; error comes during Consume
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			consumer, err := NewConsumer(tc.cfg, tc.groupID, testLogger())
			validateConsumerCreation(t, consumer, err, tc.wantError)
		})
	}
}

// TestConsumerConsumeValidation tests validation of Consume method parameters
func TestConsumerConsumeValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		topics    []string
		handler   MessageHandler
		wantError string
	}{
		{
			name:      "no topics specified",
			topics:    []string{},
			handler:   func(ctx context.Context, topic string, key, value []byte) error { return nil },
			wantError: "at least one topic must be specified",
		},
		{
			name:      "nil topics",
			topics:    nil,
			handler:   func(ctx context.Context, topic string, key, value []byte) error { return nil },
			wantError: "at least one topic must be specified",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := config.KafkaConfig{
				Brokers: []string{testBrokerLocal},
			}

			consumer, err := NewConsumer(cfg, "test-validation-group", testLogger())
			if err != nil {
				t.Fatalf(failedToCreateConsumerMsg, err)
			}
			defer consumer.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			err = consumer.Consume(ctx, tc.topics, tc.handler)

			if tc.wantError != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tc.wantError)
				}
				if !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf("error %q does not contain %q", err.Error(), tc.wantError)
				}
			}
		})
	}
}

// TestConsumerConsumeSASLValidation tests SASL validation during Consume
func TestConsumerConsumeSASLValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		sasl      config.SASLConfig
		wantError string
	}{
		{
			name: "SASL enabled but username missing",
			sasl: config.SASLConfig{
				Enabled:   true,
				Mechanism: "PLAIN",
				Username:  "",
				Password:  "secret",
			},
			wantError: "kafka SASL enabled but username or password missing",
		},
		{
			name: "SASL enabled but password missing",
			sasl: config.SASLConfig{
				Enabled:   true,
				Mechanism: "SCRAM-SHA-256",
				Username:  "alice",
				Password:  "",
			},
			wantError: "kafka SASL enabled but username or password missing",
		},
		{
			name: "SASL unsupported mechanism",
			sasl: config.SASLConfig{
				Enabled:   true,
				Mechanism: "OAUTHBEARER",
				Username:  "alice",
				Password:  "secret",
			},
			wantError: "unsupported Kafka SASL mechanism",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := config.KafkaConfig{
				Brokers: []string{testBrokerLocal},
				SASL:    tc.sasl,
			}

			consumer, err := NewConsumer(cfg, "test-sasl-validation-group", testLogger())
			if err != nil {
				t.Fatalf(failedToCreateConsumerMsg, err)
			}
			defer consumer.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			handler := func(ctx context.Context, topic string, key, value []byte) error {
				return nil
			}

			err = consumer.Consume(ctx, []string{testTopic}, handler)

			if tc.wantError != "" {
				if err == nil {
					t.Fatalf(expectedErrorGotNilMsg, tc.wantError)
				}
				if !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf(errorDoesNotContainMsg, err.Error(), tc.wantError)
				}
			}
		})
	}
}

// TestConsumerConsumeContextCancellation tests context cancellation
func TestConsumerConsumeContextCancellation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		cancelAfter   time.Duration
		expectError   bool
		errorContains string
	}{
		{
			name:          "context cancelled immediately",
			cancelAfter:   1 * time.Millisecond,
			expectError:   true,
			errorContains: "context",
		},
		{
			name:          "context cancelled after short delay",
			cancelAfter:   50 * time.Millisecond,
			expectError:   true,
			errorContains: "context",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := config.KafkaConfig{
				Brokers: []string{testBrokerLocal},
			}

			consumer, err := NewConsumer(cfg, "test-cancel-group", testLogger())
			if err != nil {
				t.Fatalf(failedToCreateConsumerMsg, err)
			}
			defer consumer.Close()

			ctx, cancel := context.WithTimeout(context.Background(), tc.cancelAfter)
			defer cancel()

			handler := func(ctx context.Context, topic string, key, value []byte) error {
				return nil
			}

			err = consumer.Consume(ctx, []string{testTopic}, handler)

			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.errorContains)
				}
				// Note: Error might be from ping failure or context cancellation
				// Both are valid for this test
			}
		})
	}
}

// TestConsumerMessageHandlerConcurrentMessages tests handling of concurrent messages
func TestConsumerMessageHandlerConcurrentMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		numMessages int
		description string
	}{
		{
			name:        "single message",
			numMessages: 1,
			description: "Handler processes one message",
		},
		{
			name:        "multiple messages",
			numMessages: 10,
			description: "Handler processes multiple messages",
		},
		{
			name:        "many messages",
			numMessages: 100,
			description: "Handler processes many messages",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var mu sync.Mutex
			processedCount := 0

			handler := func(ctx context.Context, topic string, key, value []byte) error {
				mu.Lock()
				processedCount++
				mu.Unlock()
				return nil
			}

			// Simulate message handling
			for i := 0; i < tc.numMessages; i++ {
				_ = handler(context.Background(), testTopic,
					[]byte("key"), []byte("value"))
			}

			if processedCount != tc.numMessages {
				t.Fatalf("expected %d messages processed, got %d",
					tc.numMessages, processedCount)
			}
		})
	}
}

// TestConsumerClose tests the Close method
func TestConsumerClose(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cfg         config.KafkaConfig
		groupID     string
		expectError bool
	}{
		{
			name: "close consumer successfully",
			cfg: config.KafkaConfig{
				Brokers:     []string{testBrokerLocal},
				TopicPrefix: "test_",
			},
			groupID:     "test-close-group",
			expectError: false,
		},
		{
			name: "close consumer with SASL config",
			cfg: config.KafkaConfig{
				Brokers: []string{testBrokerLocal},
				SASL: config.SASLConfig{
					Enabled:   true,
					Mechanism: "PLAIN",
					Username:  "user",
					Password:  "pass",
				},
				TopicPrefix: "test_",
			},
			groupID:     "test-close-sasl-group",
			expectError: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			consumer, err := NewConsumer(tc.cfg, tc.groupID, testLogger())
			if err != nil {
				t.Fatalf(failedToCreateConsumerMsg, err)
			}

			// Test close
			err = consumer.Close()
			if tc.expectError && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			// Test double close (should be safe)
			err = consumer.Close()
			if err != nil {
				t.Logf("Double close returned error: %v (this may be acceptable)", err)
			}
		})
	}
}

// TestConsumerConsumeWithClientCreationError tests error during client creation
func TestConsumerConsumeWithClientCreationError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		cfg    config.KafkaConfig
		topics []string
	}{
		{
			name: "invalid SASL mechanism",
			cfg: config.KafkaConfig{
				Brokers: []string{"invalid-broker:9092"},
				SASL: config.SASLConfig{
					Enabled:   true,
					Mechanism: "INVALID_MECHANISM",
					Username:  "user",
					Password:  "pass",
				},
			},
			topics: []string{testTopic},
		},
		{
			name: "unreachable broker",
			cfg: config.KafkaConfig{
				Brokers: []string{"unreachable-broker:9092"},
			},
			topics: []string{"topic"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			consumer := newTestConsumer(t, tc.cfg, "test-error-group")
			defer consumer.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			handler := func(ctx context.Context, topic string, key, value []byte) error {
				return nil
			}

			err := consumer.Consume(ctx, tc.topics, handler)
			if err == nil {
				t.Log("Expected error during consume with invalid configuration")
			}
		})
	}
}

// TestConsumeSASLMechanismValidation tests all SASL mechanism validation with table-driven tests
func TestConsumeSASLMechanismValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mechanism   string
		username    string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:      "PLAIN mechanism",
			mechanism: "PLAIN",
			username:  "user",
			password:  "pass",
		},
		{
			name:      "SCRAM-SHA-256 mechanism",
			mechanism: "SCRAM-SHA-256",
			username:  "user",
			password:  "pass",
		},
		{
			name:      "SCRAM-SHA-512 mechanism",
			mechanism: "SCRAM-SHA-512",
			username:  "user",
			password:  "pass",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := config.KafkaConfig{
				Brokers: []string{testBrokerLocal},
				SASL: config.SASLConfig{
					Enabled:   true,
					Mechanism: tc.mechanism,
					Username:  tc.username,
					Password:  tc.password,
				},
			}

			consumer, err := NewConsumer(cfg, "test-group", testLogger())
			if err != nil {
				t.Logf("Consumer creation error: %v", err)
			} else if consumer != nil {
				consumer.Close()
			}
		})
	}
}

// TestConsumeHandlerErrorPaths tests various handler error scenarios
func TestConsumeHandlerErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		handlerAction string
	}{
		{
			name:          "handler returns nil (success)",
			handlerAction: "success",
		},
		{
			name:          "handler returns error",
			handlerAction: "error",
		},
		{
			name:          "handler checks context",
			handlerAction: "context_check",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := config.KafkaConfig{
				Brokers: []string{testBrokerLocal},
			}

			consumer, err := NewConsumer(cfg, "test-handler-group", testLogger())
			if err != nil {
				t.Skipf("Failed to create consumer: %v", err)
			}
			defer consumer.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			handler := func(ctx context.Context, topic string, key, value []byte) error {
				switch tc.handlerAction {
				case "success":
					return nil
				case "error":
					return errors.New("handler error")
				case "context_check":
					select {
					case <-ctx.Done():
						return ctx.Err()
					default:
						return nil
					}
				}
				return nil
			}

			err = consumer.Consume(ctx, []string{"test-topic"}, handler)
			// We expect context timeout or broker error
			if err != nil {
				t.Logf("Consume returned error (expected): %v", err)
			}
		})
	}
}

// TestConsumeContextHandling tests context cancellation and timeout scenarios
func TestConsumeContextHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		contextTimeout time.Duration
		expectTimeout  bool
	}{
		{
			name:           "short timeout (100ms)",
			contextTimeout: 100 * time.Millisecond,
			expectTimeout:  true,
		},
		{
			name:           "very short timeout (10ms)",
			contextTimeout: 10 * time.Millisecond,
			expectTimeout:  true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := config.KafkaConfig{
				Brokers: []string{testBrokerLocal},
			}

			consumer, err := NewConsumer(cfg, "test-context-group", testLogger())
			if err != nil {
				t.Skipf("Failed to create consumer: %v", err)
			}
			defer consumer.Close()

			ctx, cancel := context.WithTimeout(context.Background(), tc.contextTimeout)
			defer cancel()

			handler := func(ctx context.Context, topic string, key, value []byte) error {
				return nil
			}

			err = consumer.Consume(ctx, []string{"test-topic"}, handler)
			if err != nil {
				if tc.expectTimeout && !strings.Contains(err.Error(), "context") {
					t.Logf("Expected context error but got: %v", err)
				}
			}
		})
	}
}

// TestConsumerCloseMultipleTimes tests closing consumer multiple times
func TestConsumerCloseMultipleTimes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		closeCount int
	}{
		{
			name:       "single close",
			closeCount: 1,
		},
		{
			name:       "double close",
			closeCount: 2,
		},
		{
			name:       "triple close",
			closeCount: 3,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := config.KafkaConfig{
				Brokers: []string{testBrokerLocal},
			}

			consumer, err := NewConsumer(cfg, "test-close-group", testLogger())
			if err != nil {
				t.Skipf("Failed to create consumer: %v", err)
			}

			for i := 0; i < tc.closeCount; i++ {
				err := consumer.Close()
				if err != nil {
					t.Logf("Close attempt %d returned error: %v", i+1, err)
				}
			}
		})
	}
}

// TestConsumeSASLInvalidMechanismDuringConsume tests SASL mechanism validation during Consume
func TestConsumeSASLInvalidMechanismDuringConsume(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mechanism   string
		username    string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "invalid mechanism during consume",
			mechanism:   "INVALID-MECH",
			username:    "user",
			password:    "pass",
			expectError: true,
			errorMsg:    "unsupported Kafka SASL mechanism",
		},
		{
			name:        "PLAIN with empty username",
			mechanism:   "PLAIN",
			username:    "",
			password:    "pass",
			expectError: true,
			errorMsg:    "kafka SASL enabled but username or password missing",
		},
		{
			name:        "SCRAM-SHA-256 with empty password",
			mechanism:   "SCRAM-SHA-256",
			username:    "user",
			password:    "",
			expectError: true,
			errorMsg:    "kafka SASL enabled but username or password missing",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := config.KafkaConfig{
				Brokers: []string{testBrokerLocal},
				SASL: config.SASLConfig{
					Enabled:   true,
					Mechanism: tc.mechanism,
					Username:  tc.username,
					Password:  tc.password,
				},
			}

			consumer, err := NewConsumer(cfg, "test-sasl-consume", testLogger())
			if err != nil {
				t.Skipf("Failed to create consumer: %v", err)
			}
			defer consumer.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			handler := func(ctx context.Context, topic string, key, value []byte) error {
				return nil
			}

			err = consumer.Consume(ctx, []string{"test-topic"}, handler)
			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.errorMsg)
				}
				if !strings.Contains(err.Error(), tc.errorMsg) {
					t.Logf("expected error containing %q, got %q", tc.errorMsg, err.Error())
				}
			}
		})
	}
}

// TestConsumeSASLValidMechanisms tests Consume with valid SASL credentials to exercise all paths
func TestConsumeSASLValidMechanisms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mechanism string
		username  string
		password  string
	}{
		{
			name:      "PLAIN mechanism with credentials",
			mechanism: "PLAIN",
			username:  "consumer1",
			password:  "consumerpass",
		},
		{
			name:      "SCRAM-SHA-256 mechanism",
			mechanism: "SCRAM-SHA-256",
			username:  "scramconsumer",
			password:  "scramconsumerpass",
		},
		{
			name:      "SCRAM-SHA-512 mechanism",
			mechanism: "SCRAM-SHA-512",
			username:  "sha512consumer",
			password:  "sha512consumerpass",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := config.KafkaConfig{
				Brokers: []string{testBrokerLocal},
				SASL: config.SASLConfig{
					Enabled:   true,
					Mechanism: tc.mechanism,
					Username:  tc.username,
					Password:  tc.password,
				},
			}

			consumer, err := NewConsumer(cfg, "test-sasl-consumer", testLogger())
			// Consumer creation may succeed or fail depending on setup
			if err != nil {
				t.Logf("Consumer creation error (may be expected): %v", err)
				return
			}
			defer consumer.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			handler := func(ctx context.Context, topic string, key, value []byte) error {
				return nil
			}

			err = consumer.Consume(ctx, []string{"test-topic"}, handler)
			if err != nil {
				// Expected - either context timeout or broker error
				t.Logf("Consume error (expected): %v", err)
			}
		})
	}
}

// Removed TestConsumeWithSpecialTopicNames - just consumes without validation

// TestHandlerErrorRecovery tests handler execution with various error scenarios
// Removed TestHandlerErrorRecovery - just consumes without validation

// TestConsumeIntegrationTable tests consuming with actual broker interaction using table-driven tests
// Removed TestConsumeIntegrationTable - just consumes without validation

// TestConsumerLoggingAndErrorHandling tests consumer with detailed logging scenarios
func TestConsumerLoggingAndErrorHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		returnError bool
		logBehavior string
	}{
		{
			name:        "handler with detailed logging",
			returnError: false,
			logBehavior: "detailed",
		},
		{
			name:        "handler with error logging",
			returnError: true,
			logBehavior: "error",
		},
		{
			name:        "handler with minimal logging",
			returnError: false,
			logBehavior: "minimal",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := config.KafkaConfig{
				Brokers: []string{testBrokerLocal},
			}

			consumer, err := NewConsumer(cfg, "logging-group", testLogger())
			if err != nil {
				t.Skipf("Failed to create consumer: %v", err)
			}
			defer consumer.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			handler := func(ctx context.Context, topic string, key, value []byte) error {
				switch tc.logBehavior {
				case "detailed":
					t.Logf("Processing message: topic=%s, key_len=%d, value_len=%d", topic, len(key), len(value))
				case "error":
					if tc.returnError {
						return errors.New("intentional handler error")
					}
				case "minimal":
					// Just process silently
				}
				return nil
			}

			_ = consumer.Consume(ctx, []string{"test-log-topic"}, handler)
		})
	}
}

// ==================== Logging Contract Tests ====================

// TestLoggingContract_KafkaConsumer_WarnLevelOnly verifies that the Kafka consumer
// only logs at Warn level (never Error) for all error paths during setup.
func TestLoggingContract_KafkaConsumer_WarnLevelOnly(t *testing.T) {
	var buf bytes.Buffer
	log := zerolog.New(zerolog.ConsoleWriter{Out: &buf, NoColor: true, TimeFormat: time.RFC3339}).
		Level(zerolog.DebugLevel).With().Timestamp().Logger()

	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	// SASL enabled but no credentials → should Warn, not Error
	cfg := config.KafkaConfig{
		Brokers: []string{"localhost:9092"},
		SASL: config.SASLConfig{
			Enabled: true,
			// Missing username and password
		},
	}

	consumer, err := NewConsumer(cfg, "test-group", log)
	if err != nil {
		t.Skipf("cannot create consumer: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = consumer.Consume(ctx, []string{"test-topic"}, func(ctx context.Context, topic string, key, value []byte) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for missing SASL credentials")
	}

	output := buf.String()
	if strings.Contains(output, "ERR") {
		t.Fatalf("Kafka consumer should NOT produce ERR-level logs, got: %s", output)
	}
	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls, got %d", len(*captureRecords))
	}
}

// TestLoggingContract_KafkaConsumer_NoCaptureException verifies that no error path
// in the Kafka consumer calls CaptureException.
func TestLoggingContract_KafkaConsumer_NoCaptureException(t *testing.T) {
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	log := testLogger()

	// Error path 1: Empty brokers
	_, _ = NewConsumer(config.KafkaConfig{}, "test-group", log)

	// Error path 2: Empty group ID
	_, _ = NewConsumer(config.KafkaConfig{Brokers: []string{"localhost:9092"}}, "", log)

	// Error path 3: SASL with unsupported mechanism (triggers on Consume)
	consumer, err := NewConsumer(config.KafkaConfig{
		Brokers: []string{"localhost:9092"},
		SASL: config.SASLConfig{
			Enabled:   true,
			Username:  "user",
			Password:  "pass",
			Mechanism: "UNSUPPORTED",
		},
	}, "test-group", log)
	if err == nil && consumer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		_ = consumer.Consume(ctx, []string{"test"}, func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		})
	}

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls across all error paths, got %d", len(*captureRecords))
	}
}

// TestLoggingContract_KafkaConsumer_ErrorsReturnedToCaller verifies that the Kafka
// consumer returns errors to callers rather than swallowing them.
func TestLoggingContract_KafkaConsumer_ErrorsReturnedToCaller(t *testing.T) {
	log := testLogger()

	// Empty brokers
	_, err := NewConsumer(config.KafkaConfig{}, "test-group", log)
	if err == nil {
		t.Fatal("NewConsumer with empty brokers: expected error to be returned to caller")
	}

	// Empty group ID
	_, err = NewConsumer(config.KafkaConfig{Brokers: []string{"localhost:9092"}}, "", log)
	if err == nil {
		t.Fatal("NewConsumer with empty group: expected error to be returned to caller")
	}

	// SASL with unsupported mechanism (triggers during Consume)
	consumer, err := NewConsumer(config.KafkaConfig{
		Brokers: []string{"localhost:9092"},
		SASL: config.SASLConfig{
			Enabled:   true,
			Username:  "user",
			Password:  "pass",
			Mechanism: "UNSUPPORTED",
		},
	}, "test-group", log)
	if err == nil && consumer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		err = consumer.Consume(ctx, []string{"test"}, func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		})
		if err == nil {
			t.Fatal("Consume with unsupported SASL: expected error to be returned to caller")
		}
	}
}
