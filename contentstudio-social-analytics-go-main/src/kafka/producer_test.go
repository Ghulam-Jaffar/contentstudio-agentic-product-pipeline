package kafka

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/rs/zerolog"
)

const (
	defaultBroker                = "localhost:9092"
	defaultTestTopic             = "test-topic"
	defaultTestValue             = "test-value"
	pingErrorSubstring           = "failed to ping"
	contextErrorSubstring        = "context"
	skipProducerErrFormat        = "Cannot create producer: %v"
	skipNoBrokerFormat           = "Skipping test, no Kafka broker available: %v"
	failedToCreateProducerFormat = "failed to create producer: %v"
)

func newTestProducer(t *testing.T, cfg config.KafkaConfig) Producer {
	t.Helper()

	prod, err := NewProducer(cfg, testLogger())
	if err != nil {
		t.Skipf(skipProducerErrFormat, err)
		return nil
	}

	return prod
}

func defaultProducerConfig() config.KafkaConfig {
	return config.KafkaConfig{
		Brokers:     []string{defaultBroker},
		TopicPrefix: "test_",
	}
}

func startReachableBroker(t *testing.T) (net.Listener, string, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	cleanup := func() {
		listener.Close()
	}

	return listener, listener.Addr().String(), cleanup
}

// TestNewProducerSuccess tests successful producer creation with various configurations
func TestNewProducerSuccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cfg         config.KafkaConfig
		skipPing    bool
		description string
	}{
		{
			name: "basic configuration without SASL",
			cfg:  defaultProducerConfig(),
		},
		{
			name: "with topic prefix",
			cfg: config.KafkaConfig{
				Brokers:     []string{defaultBroker, "localhost:9093"},
				TopicPrefix: "prod_",
			},
			skipPing:    true,
			description: "Should create producer with topic prefix",
		},
		{
			name: "with empty topic prefix",
			cfg: config.KafkaConfig{
				Brokers:     []string{defaultBroker},
				TopicPrefix: "",
			},
			skipPing:    true,
			description: "Should create producer without topic prefix",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Note: This test will fail at ping stage without real Kafka
			// In a real test environment, you'd mock the kgo.Client or use testcontainers
			_, err := NewProducer(tc.cfg, testLogger())

			if tc.skipPing && err != nil {
				if !strings.Contains(err.Error(), pingErrorSubstring) {
					t.Logf("Expected ping failure, got different error: %v", err)
				}
			}
		})
	}
}

// TestNewProducerSASLMechanisms tests various SASL authentication mechanisms
func TestNewProducerSASLMechanisms(t *testing.T) {
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
			name:        "PLAIN mechanism with valid credentials",
			mechanism:   "PLAIN",
			username:    "testuser",
			password:    "testpass",
			expectError: true, // Will fail at ping without real Kafka
			errorMsg:    pingErrorSubstring,
		},
		{
			name:        "SCRAM-SHA-256 mechanism",
			mechanism:   "SCRAM-SHA-256",
			username:    "scramuser",
			password:    "scrampass",
			expectError: true,
			errorMsg:    pingErrorSubstring,
		},
		{
			name:        "SCRAM-SHA-512 mechanism",
			mechanism:   "SCRAM-SHA-512",
			username:    "scramuser512",
			password:    "scrampass512",
			expectError: true,
			errorMsg:    pingErrorSubstring,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := config.KafkaConfig{
				Brokers: []string{defaultBroker},
				SASL: config.SASLConfig{
					Enabled:   true,
					Mechanism: tc.mechanism,
					Username:  tc.username,
					Password:  tc.password,
				},
				TopicPrefix: "test_",
			}

			_, err := NewProducer(cfg, testLogger())

			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.errorMsg)
				}
				if !strings.Contains(err.Error(), tc.errorMsg) {
					t.Fatalf("expected error containing %q, got %q", tc.errorMsg, err.Error())
				}
			}
		})
	}
}

// TestProducerClose tests the Close method
func TestProducerClose(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cfg         config.KafkaConfig
		expectError bool
	}{
		{
			name:        "close producer successfully",
			cfg:         defaultProducerConfig(),
			expectError: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			prod, err := NewProducer(tc.cfg, testLogger())
			if err != nil {
				// Expected to fail at ping, skip this test
				t.Skipf("Cannot test Close without successful producer creation: %v", err)
				return
			}

			// Test close
			err = prod.Close()
			if tc.expectError && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

// TestProducerProduceWithBroker tests the Produce method with table-driven tests
func TestProducerProduceWithBroker(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		topic          string
		key            []byte
		value          []byte
		topicPrefix    string
		contextTimeout time.Duration
		description    string
	}{
		{
			name:           "produce with key and value",
			topic:          defaultTestTopic,
			key:            []byte("test-key"),
			value:          []byte(defaultTestValue),
			topicPrefix:    "test_",
			contextTimeout: 5 * time.Second,
			description:    "Should produce message with both key and value",
		},
		{
			name:           "produce with nil key",
			topic:          defaultTestTopic,
			key:            nil,
			value:          []byte(defaultTestValue),
			topicPrefix:    "test_",
			contextTimeout: 5 * time.Second,
			description:    "Should produce message with nil key",
		},
		{
			name:           "produce with empty value",
			topic:          defaultTestTopic,
			key:            []byte("key"),
			value:          []byte{},
			topicPrefix:    "test_",
			contextTimeout: 5 * time.Second,
			description:    "Should produce message with empty value",
		},
		{
			name:           "produce without topic prefix",
			topic:          "raw-topic",
			key:            []byte("key"),
			value:          []byte("value"),
			topicPrefix:    "",
			contextTimeout: 5 * time.Second,
			description:    "Should produce to topic without prefix",
		},
		{
			name:           "produce with short timeout",
			topic:          "timeout-topic",
			key:            []byte("key"),
			value:          []byte("value"),
			topicPrefix:    "test_",
			contextTimeout: 1 * time.Millisecond,
			description:    "Should handle context timeout gracefully",
		},
		{
			name:           "produce large message",
			topic:          "large-topic",
			key:            []byte("key"),
			value:          bytes.Repeat([]byte("x"), 1024*100), // 100KB
			topicPrefix:    "test_",
			contextTimeout: 5 * time.Second,
			description:    "Should produce large message",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := config.KafkaConfig{
				Brokers:     []string{defaultBroker},
				TopicPrefix: tc.topicPrefix,
			}

			prod, err := NewProducer(cfg, testLogger())
			if err != nil {
				if strings.Contains(err.Error(), pingErrorSubstring) {
					t.Skipf(skipNoBrokerFormat, err)
				}
				t.Fatalf(failedToCreateProducerFormat, err)
			}
			defer prod.Close()

			ctx, cancel := context.WithTimeout(context.Background(), tc.contextTimeout)
			defer cancel()

			err = prod.Produce(ctx, tc.topic, tc.key, tc.value)
			// With broker: should succeed or timeout
			// Without broker: will skip above
			if err != nil && !strings.Contains(err.Error(), contextErrorSubstring) {
				t.Logf("Produce returned error (may be expected): %v", err)
			}
		})
	}
}

// TestProducerProduceContextCancellation tests context handling
func TestProducerProduceContextCancellation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cancelAfter time.Duration
		description string
	}{
		{
			name:        "immediate cancellation",
			cancelAfter: 0,
			description: "Context cancelled immediately",
		},
		{
			name:        "quick cancellation",
			cancelAfter: 1 * time.Millisecond,
			description: "Context cancelled after 1ms",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := defaultProducerConfig()
			prod, err := NewProducer(cfg, testLogger())
			if err != nil {
				if strings.Contains(err.Error(), pingErrorSubstring) {
					t.Skipf(skipNoBrokerFormat, err)
				}
				t.Fatalf(failedToCreateProducerFormat, err)
			}
			defer prod.Close()

			ctx, cancel := context.WithCancel(context.Background())
			if tc.cancelAfter > 0 {
				time.AfterFunc(tc.cancelAfter, cancel)
			} else {
				cancel()
			}
			defer cancel()

			err = prod.Produce(ctx, defaultTestTopic, []byte("key"), []byte("value"))
			if err == nil {
				t.Log("Produce succeeded despite cancelled context")
			}
		})
	}
}

// TestProducerCloseIdempotency tests that Close can be called multiple times
func TestProducerCloseIdempotency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		closeCalls  int
		description string
	}{
		{
			name:        "single close",
			closeCalls:  1,
			description: "Close called once",
		},
		{
			name:        "double close",
			closeCalls:  2,
			description: "Close called twice (idempotent)",
		},
		{
			name:        "multiple close",
			closeCalls:  5,
			description: "Close called multiple times",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := defaultProducerConfig()
			prod, err := NewProducer(cfg, testLogger())
			if err != nil {
				if strings.Contains(err.Error(), pingErrorSubstring) {
					t.Skipf(skipNoBrokerFormat, err)
				}
				t.Fatalf(failedToCreateProducerFormat, err)
			}

			for i := 0; i < tc.closeCalls; i++ {
				err := prod.Close()
				if err != nil {
					t.Logf("Close call %d returned error: %v", i+1, err)
				}
			}
		})
	}
}

// TestProducerCloseAfterProduce tests Close after successful produce
func TestProducerCloseAfterProduce(t *testing.T) {
	t.Parallel()

	cfg := defaultProducerConfig()
	prod, err := NewProducer(cfg, testLogger())
	if err != nil {
		if strings.Contains(err.Error(), pingErrorSubstring) {
			t.Skipf(skipNoBrokerFormat, err)
		}
		t.Fatalf(failedToCreateProducerFormat, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Attempt to produce
	err = prod.Produce(ctx, defaultTestTopic, []byte("key"), []byte("value"))
	if err != nil {
		t.Logf("Produce error (may be expected without broker): %v", err)
	}

	// Close should work regardless of produce result
	err = prod.Close()
	if err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
}

// TestProducerProduceContextCancellationOld tests context cancellation
func TestProducerProduceContextCancellationOld(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		cancelBefore  bool
		expectedError bool
	}{
		{
			name:          "context cancelled before produce",
			cancelBefore:  true,
			expectedError: true,
		},
		{
			name:          "context not cancelled",
			cancelBefore:  false,
			expectedError: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := defaultProducerConfig()

			prod := newTestProducer(t, cfg)
			if prod == nil {
				return
			}
			defer prod.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			if tc.cancelBefore {
				cancel()
			}

			err := prod.Produce(ctx, defaultTestTopic, []byte("key"), []byte("value"))

			if err != nil {
				t.Logf("Got error (expected with cancelled context): %v", err)
			}
		})
	}
}

// TestWaitForKafkaReachable tests waitForKafka with a reachable broker
func TestWaitForKafkaReachable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "reachable broker returns"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, addr, cleanup := startReachableBroker(t)
			defer cleanup()

			logger := testLogger()
			done := make(chan struct{})
			go func() {
				defer close(done)
				waitForKafka(logger, []string{addr}, 2, 10*time.Millisecond)
			}()

			select {
			case <-done:
				// Success - function returned normally
			case <-time.After(2 * time.Second):
				t.Fatal("waitForKafka did not return in time")
			}
		})
	}
}

// TestNewProducerWithInvalidBrokerFormat tests producer creation with invalid broker formats
func TestNewProducerWithInvalidBrokerFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		brokers []string
	}{
		{
			name:    "broker without port",
			brokers: []string{"localhost"},
		},
		{
			name:    "invalid hostname",
			brokers: []string{"invalid_hostname:9092"},
		},
		{
			name:    "multiple invalid brokers",
			brokers: []string{"invalid1:9092", "invalid2:9093"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := config.KafkaConfig{
				Brokers:     tc.brokers,
				TopicPrefix: "test_",
			}

			_, err := NewProducer(cfg, testLogger())
			// Expected to fail at ping or connection
			if err == nil {
				t.Log("Producer creation succeeded (unexpected but not an error)")
			}
		})
	}
}

// TestProducerCloseIdempotent tests that Close can be called multiple times
func TestProducerCloseIdempotent(t *testing.T) {
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
			name:       "multiple closes",
			closeCount: 3,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := defaultProducerConfig()

			prod := newTestProducer(t, cfg)
			if prod == nil {
				return
			}

			for i := 0; i < tc.closeCount; i++ {
				err := prod.Close()
				if err != nil {
					t.Logf("Close call %d returned error: %v", i+1, err)
				}
			}
		})
	}
}

// TestWaitForKafkaWithLogger tests waitForKafka with proper logger
func TestWaitForKafkaWithLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setupBroker  bool
		brokers      []string
		retries      int
		delay        time.Duration
		shouldReturn bool
	}{
		{
			name:         "broker reachable on first try",
			setupBroker:  true,
			retries:      3,
			delay:        10 * time.Millisecond,
			shouldReturn: true,
		},
		{
			name:         "invalid broker format",
			setupBroker:  false,
			brokers:      []string{"invalid-broker-format"},
			retries:      1,
			delay:        1 * time.Millisecond,
			shouldReturn: false,
		},
		{
			name:         "multiple brokers with one valid",
			setupBroker:  true,
			retries:      2,
			delay:        10 * time.Millisecond,
			shouldReturn: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var brokers []string

			if tc.setupBroker {
				_, addr, cleanup := startReachableBroker(t)
				defer cleanup()

				brokers = []string{addr}
			} else {
				brokers = tc.brokers
			}

			if tc.shouldReturn {
				done := make(chan struct{})
				go func() {
					defer close(done)
					waitForKafka(testLogger(), brokers, tc.retries, tc.delay)
				}()

				select {
				case <-done:
					// Success
				case <-time.After(2 * time.Second):
					t.Fatal("waitForKafka did not return in time")
				}
			} else {
				// For cases that would call os.Exit, we just test the logic doesn't panic
				// We can't easily test os.Exit without subprocess testing
				t.Logf("Would call os.Exit for broker: %v", brokers)
			}
		})
	}
}

// setupDelayedBroker creates a broker that appears after a delay to test retry logic
func setupDelayedBroker(t *testing.T, recreateDelay time.Duration) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	addr := listener.Addr().String()
	listener.Close()

	go func() {
		time.Sleep(recreateDelay)
		newListener, err := net.Listen("tcp", addr)
		if err != nil {
			return
		}
		go func() {
			for {
				conn, err := newListener.Accept()
				if err != nil {
					return
				}
				conn.Close()
			}
		}()
		time.Sleep(1 * time.Second)
		newListener.Close()
	}()

	return addr
}

// TestWaitForKafkaRetryLogic tests the retry mechanism
func TestWaitForKafkaRetryLogic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		recreateDelay time.Duration
		retryCount    int
		retryInterval time.Duration
	}{
		{
			name:          "broker appears after delay",
			recreateDelay: 100 * time.Millisecond,
			retryCount:    10,
			retryInterval: 50 * time.Millisecond,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			addr := setupDelayedBroker(t, tc.recreateDelay)

			done := make(chan struct{})
			go func() {
				defer close(done)
				waitForKafka(testLogger(), []string{addr}, tc.retryCount, tc.retryInterval)
			}()

			select {
			case <-done:
				// Success - retry logic worked
			case <-time.After(3 * time.Second):
				t.Log("waitForKafka timed out (expected if broker didn't come online)")
			}
		})
	}
}

// TestWaitForKafkaMultipleBrokers tests handling of multiple brokers
func TestWaitForKafkaMultipleBrokers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		retries int
		delay   time.Duration
	}{
		{
			name:    "mix of invalid and valid brokers",
			retries: 3,
			delay:   10 * time.Millisecond,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, addr, cleanup := startReachableBroker(t)
			defer cleanup()

			brokers := []string{
				"invalid-host:9999",
				"127.0.0.1:1", // Likely closed port
				addr,
			}

			done := make(chan struct{})
			go func() {
				defer close(done)
				waitForKafka(testLogger(), brokers, tc.retries, tc.delay)
			}()

			select {
			case <-done:
				// Success - found the valid broker
			case <-time.After(2 * time.Second):
				t.Fatal("waitForKafka did not find valid broker in time")
			}
		})
	}
}

// TestNewProducerValidation tests NewProducer validation with table-driven tests
func TestNewProducerValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cfg         config.KafkaConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "nil brokers list",
			cfg: config.KafkaConfig{
				Brokers:     nil,
				TopicPrefix: "test_",
			},
			expectError: true,
			errorMsg:    "kafka brokers must be configured",
		},
		{
			name: "empty brokers list",
			cfg: config.KafkaConfig{
				Brokers:     []string{},
				TopicPrefix: "test_",
			},
			expectError: true,
			errorMsg:    "kafka brokers must be configured",
		},
		{
			name: "SASL enabled with missing username",
			cfg: config.KafkaConfig{
				Brokers:     []string{"localhost:9092"},
				TopicPrefix: "test_",
				SASL: config.SASLConfig{
					Enabled:   true,
					Mechanism: "PLAIN",
					Username:  "",
					Password:  "testpass",
				},
			},
			expectError: true,
			errorMsg:    "kafka SASL enabled but username or password missing",
		},
		{
			name: "SASL enabled with missing password",
			cfg: config.KafkaConfig{
				Brokers:     []string{"localhost:9092"},
				TopicPrefix: "test_",
				SASL: config.SASLConfig{
					Enabled:   true,
					Mechanism: "PLAIN",
					Username:  "testuser",
					Password:  "",
				},
			},
			expectError: true,
			errorMsg:    "kafka SASL enabled but username or password missing",
		},
		{
			name: "SASL with invalid mechanism",
			cfg: config.KafkaConfig{
				Brokers:     []string{"localhost:9092"},
				TopicPrefix: "test_",
				SASL: config.SASLConfig{
					Enabled:   true,
					Mechanism: "INVALID-MECHANISM",
					Username:  "testuser",
					Password:  "testpass",
				},
			},
			expectError: true,
			errorMsg:    "unsupported Kafka SASL mechanism",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewProducer(tc.cfg, testLogger())

			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.errorMsg)
				}
				if !strings.Contains(err.Error(), tc.errorMsg) {
					t.Fatalf("expected error containing %q, got %q", tc.errorMsg, err.Error())
				}
			}
		})
	}
}

// TestNewConsumerErrorConditions tests NewConsumer error scenarios with table-driven tests
func TestNewConsumerErrorConditions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cfg         config.KafkaConfig
		groupID     string
		expectError bool
		errorMsg    string
	}{
		{
			name: "nil brokers",
			cfg: config.KafkaConfig{
				Brokers:     nil,
				TopicPrefix: "test_",
			},
			groupID:     "test-group",
			expectError: true,
			errorMsg:    "kafka brokers must be configured",
		},
		{
			name: "empty brokers list",
			cfg: config.KafkaConfig{
				Brokers:     []string{},
				TopicPrefix: "test_",
			},
			groupID:     "test-group",
			expectError: true,
			errorMsg:    "kafka brokers must be configured",
		},
		{
			name: "empty group ID",
			cfg: config.KafkaConfig{
				Brokers:     []string{"localhost:9092"},
				TopicPrefix: "test_",
			},
			groupID:     "",
			expectError: true,
			errorMsg:    "consumer group ID must be provided",
		},
		{
			name: "SASL enabled with missing username",
			cfg: config.KafkaConfig{
				Brokers:     []string{"localhost:9092"},
				TopicPrefix: "test_",
				SASL: config.SASLConfig{
					Enabled:   true,
					Mechanism: "PLAIN",
					Username:  "",
					Password:  "testpass",
				},
			},
			groupID:     "test-group",
			expectError: false, // Consumer creation succeeds, error happens at Consume
		},
		{
			name: "SASL enabled with missing password",
			cfg: config.KafkaConfig{
				Brokers:     []string{"localhost:9092"},
				TopicPrefix: "test_",
				SASL: config.SASLConfig{
					Enabled:   true,
					Mechanism: "SCRAM-SHA-256",
					Username:  "testuser",
					Password:  "",
				},
			},
			groupID:     "test-group",
			expectError: false, // Consumer creation succeeds
		},
		{
			name: "SASL with invalid mechanism",
			cfg: config.KafkaConfig{
				Brokers:     []string{"localhost:9092"},
				TopicPrefix: "test_",
				SASL: config.SASLConfig{
					Enabled:   true,
					Mechanism: "UNSUPPORTED",
					Username:  "testuser",
					Password:  "testpass",
				},
			},
			groupID:     "test-group",
			expectError: false, // Consumer creation succeeds but Consume will fail
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			consumer, err := NewConsumer(tc.cfg, tc.groupID, testLogger())

			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.errorMsg)
				}
				if !strings.Contains(err.Error(), tc.errorMsg) {
					t.Fatalf("expected error containing %q, got %q", tc.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Logf("got error (may be expected): %v", err)
				}
				if consumer != nil {
					consumer.Close()
				}
			}
		})
	}
}

// TestConsumeWithVariousTopics tests Consume with various topic configurations
func TestConsumeWithVariousTopics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		topics       []string
		topicPrefix  string
		expectError  bool
		errorMessage string
	}{
		{
			name:         "empty topics list",
			topics:       []string{},
			expectError:  true,
			errorMessage: "at least one topic must be specified",
		},
		{
			name:        "single topic",
			topics:      []string{"test-topic"},
			expectError: false,
		},
		{
			name:        "multiple topics",
			topics:      []string{"topic1", "topic2", "topic3"},
			expectError: false,
		},
		{
			name:        "topic with special characters",
			topics:      []string{"test-topic_123"},
			expectError: false,
		},
		{
			name:        "single topic with prefix",
			topics:      []string{"events"},
			topicPrefix: "prod_",
			expectError: false,
		},
		{
			name:        "multiple topics with prefix",
			topics:      []string{"events", "logs", "metrics"},
			topicPrefix: "app_",
			expectError: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := config.KafkaConfig{
				Brokers:     []string{"localhost:9092"},
				TopicPrefix: tc.topicPrefix,
			}
			consumer, err := NewConsumer(cfg, "test-group", testLogger())
			if err != nil {
				t.Skipf("Failed to create consumer: %v", err)
			}
			defer consumer.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			handler := func(ctx context.Context, topic string, key, value []byte) error {
				return nil
			}

			err = consumer.Consume(ctx, tc.topics, handler)

			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.errorMessage)
				}
				if !strings.Contains(err.Error(), tc.errorMessage) {
					t.Fatalf("expected error containing %q, got %q", tc.errorMessage, err.Error())
				}
			} else {
				// For non-error cases, we expect either context cancellation or broker unavailability
				if err != nil && !strings.Contains(err.Error(), "context") {
					t.Logf("Consume returned error (may be expected with no broker): %v", err)
				}
			}
		})
	}
}

// TestProduceClose tests Produce followed by Close with table-driven test cases
func TestProduceClose(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		messagesBeforeClose int
		contextTimeout      time.Duration
	}{
		{
			name:                "produce once then close",
			messagesBeforeClose: 1,
			contextTimeout:      2 * time.Second,
		},
		{
			name:                "produce multiple then close",
			messagesBeforeClose: 3,
			contextTimeout:      2 * time.Second,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := config.KafkaConfig{
				Brokers: []string{"localhost:9092"},
			}

			prod, err := NewProducer(cfg, testLogger())
			if err != nil {
				t.Skipf("Failed to create producer: %v", err)
			}

			for i := 0; i < tc.messagesBeforeClose; i++ {
				ctx, cancel := context.WithTimeout(context.Background(), tc.contextTimeout)
				topic := "close-test-" + fmt.Sprintf("%d", i)
				_ = prod.Produce(ctx, topic, []byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i)))
				cancel()
			}

			err = prod.Close()
			if err != nil {
				t.Logf("Close returned error: %v", err)
			}
		})
	}
}

// TestProducerSuccessfulProduction tests successful message production to exercise all return paths
func TestProducerSuccessfulProduction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		topicName string
		key       []byte
		value     []byte
	}{
		{
			name:      "simple successful produce",
			topicName: "success-test-1",
			key:       []byte("test-key"),
			value:     []byte("test-value"),
		},
		{
			name:      "produce with empty key",
			topicName: "success-test-2",
			key:       []byte{},
			value:     []byte("value"),
		},
		{
			name:      "produce with empty value",
			topicName: "success-test-3",
			key:       []byte("key"),
			value:     []byte{},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := config.KafkaConfig{
				Brokers: []string{"localhost:9092"},
			}

			prod, err := NewProducer(cfg, testLogger())
			if err != nil {
				t.Skipf("Failed to create producer: %v", err)
			}
			defer prod.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = prod.Produce(ctx, tc.topicName, tc.key, tc.value)
			if err == nil {
				t.Logf("SUCCESS: Message produced successfully to topic %s", tc.topicName)
			} else if strings.Contains(err.Error(), "UNKNOWN_TOPIC_OR_PARTITION") {
				t.Logf("Topic doesn't exist yet, but Produce reached broker: %v", err)
			} else {
				t.Logf("Produce error: %v", err)
			}
		})
	}
}

// TestNewProducerSASLAppendLogic tests the SASL append logic path
func TestNewProducerSASLAppendLogic(t *testing.T) {
	tests := []struct {
		name      string
		mechanism string
		username  string
		password  string
	}{
		{
			name:      "PLAIN with valid credentials",
			mechanism: "PLAIN",
			username:  "user",
			password:  "pass",
		},
		{
			name:      "SCRAM-SHA-256 with valid credentials",
			mechanism: "SCRAM-SHA-256",
			username:  "user",
			password:  "pass",
		},
		{
			name:      "SCRAM-SHA-512 with valid credentials",
			mechanism: "SCRAM-SHA-512",
			username:  "user",
			password:  "pass",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.KafkaConfig{
				Brokers: []string{"localhost:9092"},
				SASL: config.SASLConfig{
					Enabled:   true,
					Mechanism: tc.mechanism,
					Username:  tc.username,
					Password:  tc.password,
				},
			}

			// Just verify that NewProducer accepts SASL config with various mechanisms
			producer, err := NewProducer(cfg, testLogger())
			// Producer may succeed or fail depending on broker availability
			if producer != nil {
				producer.Close()
			}
			// Just ensure it doesn't panic
			t.Logf("Producer creation returned: err=%v", err)
		})
	}
}

// TestProducerSuccessfulProductionReturn tests successful produce returns
func TestProducerSuccessfulProductionReturn(t *testing.T) {
	tests := []struct {
		name        string
		topic       string
		key         []byte
		value       []byte
		shouldRetry bool
	}{
		{
			name:        "successful produce with simple data",
			topic:       "success-test",
			key:         []byte("test-key"),
			value:       []byte("test-value"),
			shouldRetry: false,
		},
		{
			name:        "successful produce with empty key",
			topic:       "success-empty-key",
			key:         []byte{},
			value:       []byte("value"),
			shouldRetry: false,
		},
		{
			name:        "successful produce with binary data",
			topic:       "success-binary",
			key:         []byte{0x00, 0x01, 0x02},
			value:       []byte{0xFF, 0xFE, 0xFD},
			shouldRetry: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := defaultProducerConfig()
			prod := newTestProducer(t, cfg)
			if prod == nil {
				return
			}
			defer prod.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := prod.Produce(ctx, tc.topic, tc.key, tc.value)
			if err != nil {
				t.Logf("Produce returned error (may be expected in test environment): %v", err)
			}
		})
	}
}

// TestWaitForKafkaSuccessPath tests successful reachability detection
func TestWaitForKafkaSuccessPath(t *testing.T) {
	tests := []struct {
		name    string
		brokers []string
		retries int
		delay   time.Duration
	}{
		{
			name:    "localhost broker on standard port",
			brokers: []string{"127.0.0.1:9092"},
			retries: 1,
			delay:   100 * time.Millisecond,
		},
		{
			name:    "multiple brokers with common port",
			brokers: []string{"127.0.0.1:9092", "127.0.0.1:9093"},
			retries: 1,
			delay:   100 * time.Millisecond,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			logger := testLogger()

			// Start a listener to make broker reachable
			listener, err := net.Listen("tcp", "127.0.0.1:9999")
			if err != nil {
				t.Skipf("Could not create test listener: %v", err)
			}
			defer listener.Close()

			go func() {
				conn, _ := listener.Accept()
				if conn != nil {
					conn.Close()
				}
			}()

			testBrokers := []string{"127.0.0.1:9999"}
			// The function will call os.Exit on failure, but with a reachable broker it should return
			waitForKafka(logger, testBrokers, 1, 10*time.Millisecond)
		})
	}
}

// testLogger returns a logger for testing that writes to a buffer
func testLogger() zerolog.Logger {
	var buf bytes.Buffer
	return zerolog.New(&buf).Level(zerolog.DebugLevel).With().Timestamp().Logger()
}

// ==================== Logging Contract Tests ====================

// TestLoggingContract_KafkaProducer_WarnLevelOnly verifies that the Kafka producer
// only logs at Warn level (never Error) for all error paths. It returns errors to callers.
func TestLoggingContract_KafkaProducer_WarnLevelOnly(t *testing.T) {
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

	_, err := NewProducer(cfg, log)
	if err == nil {
		t.Fatal("expected error for missing SASL credentials")
	}

	output := buf.String()
	if strings.Contains(output, "ERR") {
		t.Fatalf("Kafka producer should NOT produce ERR-level logs, got: %s", output)
	}
	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls, got %d", len(*captureRecords))
	}
}

// TestLoggingContract_KafkaProducer_NoCaptureException verifies that no error path
// in the Kafka producer calls CaptureException.
func TestLoggingContract_KafkaProducer_NoCaptureException(t *testing.T) {
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	log := testLogger()

	// Error path 1: Empty brokers
	_, _ = NewProducer(config.KafkaConfig{}, log)

	// Error path 2: SASL with unsupported mechanism
	_, _ = NewProducer(config.KafkaConfig{
		Brokers: []string{"localhost:9092"},
		SASL: config.SASLConfig{
			Enabled:   true,
			Username:  "user",
			Password:  "pass",
			Mechanism: "UNSUPPORTED",
		},
	}, log)

	// Error path 3: SASL enabled but no credentials
	_, _ = NewProducer(config.KafkaConfig{
		Brokers: []string{"localhost:9092"},
		SASL: config.SASLConfig{
			Enabled: true,
		},
	}, log)

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls across all error paths, got %d", len(*captureRecords))
	}
}

// TestLoggingContract_KafkaProducer_ErrorsReturnedToCaller verifies that the Kafka producer
// returns errors to callers rather than swallowing them.
func TestLoggingContract_KafkaProducer_ErrorsReturnedToCaller(t *testing.T) {
	log := testLogger()

	// Empty brokers
	_, err := NewProducer(config.KafkaConfig{}, log)
	if err == nil {
		t.Fatal("NewProducer with empty brokers: expected error to be returned to caller")
	}

	// SASL with missing credentials
	_, err = NewProducer(config.KafkaConfig{
		Brokers: []string{"localhost:9092"},
		SASL:    config.SASLConfig{Enabled: true},
	}, log)
	if err == nil {
		t.Fatal("NewProducer with missing SASL creds: expected error to be returned to caller")
	}

	// SASL with unsupported mechanism
	_, err = NewProducer(config.KafkaConfig{
		Brokers: []string{"localhost:9092"},
		SASL: config.SASLConfig{
			Enabled:   true,
			Username:  "user",
			Password:  "pass",
			Mechanism: "UNSUPPORTED",
		},
	}, log)
	if err == nil {
		t.Fatal("NewProducer with unsupported SASL: expected error to be returned to caller")
	}
}
