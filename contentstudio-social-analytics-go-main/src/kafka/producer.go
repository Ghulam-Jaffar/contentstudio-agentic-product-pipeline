package kafka

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"
	"github.com/twmb/franz-go/plugin/kzerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
)

// Producer defines the interface for a Kafka producer.
type Producer interface {
	// Produce sends a message to the specified topic.
	// The key and value are byte slices.
	Produce(ctx context.Context, topic string, key, value []byte) error
	// Close shuts down the producer and releases resources.
	Close() error
}

type franzKafkaProducer struct {
	client      *kgo.Client
	logger      zerolog.Logger
	topicPrefix string
}

// NewProducer creates a new Kafka producer instance using franz-go.
// It takes Kafka configuration and a Zerolog logger.
func NewProducer(cfg config.KafkaConfig, logger zerolog.Logger) (Producer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, errors.New("kafka brokers must be configured")
	}

	kgoClientLogger := logger.With().Str("component", "kgo_client").Logger()
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.WithLogger(kzerolog.New(&kgoClientLogger)),
		// Adjust batching and linger for a balance of latency and throughput.
		// These are just examples; tune them based on your workload.
		kgo.ProducerBatchMaxBytes(1 * 1024 * 1024), // 1MB
		kgo.ProducerLinger(100 * time.Millisecond), // Wait up to 100ms to fill a batch
		// Add a default timeout for produce requests.
		kgo.ProduceRequestTimeout(10 * time.Second),
	}

	// Example: Add SASL or TLS options if configured
	// if cfg.SASL.Enabled {
	// 	opts = append(opts, kgo.SASL(scram.Auth{User: cfg.SASL.User, Pass: cfg.SASL.Password}.AsSha256Mechanism()))
	// }
	// if cfg.TLS.Enabled {
	//    tlsConfig, err := setupTLSConfig(cfg.TLS) // Your function to build *tls.Config
	//    if err != nil {
	//        return nil, fmt.Errorf("failed to setup TLS config for Kafka: %w", err)
	//    }
	//	  opts = append(opts, kgo.DialTLSConfig(tlsConfig))
	// }

	if cfg.SASL.Enabled {
		if cfg.SASL.Username == "" || cfg.SASL.Password == "" {
			logger.Warn().Msg("Kafka SASL is enabled but username or password is not configured")
			return nil, errors.New("kafka SASL enabled but username or password missing")
		}
		var saslOpt kgo.Opt
		switch cfg.SASL.Mechanism {
		case "PLAIN":
			saslOpt = kgo.SASL(plain.Auth{
				User: cfg.SASL.Username,
				Pass: cfg.SASL.Password,
			}.AsMechanism())
			logger.Info().Str("mechanism", "PLAIN").Msg("Configuring Kafka SASL PLAIN authentication")
		case "SCRAM-SHA-256":
			saslOpt = kgo.SASL(scram.Auth{
				User: cfg.SASL.Username,
				Pass: cfg.SASL.Password,
			}.AsSha256Mechanism())
			logger.Info().Str("mechanism", "SCRAM-SHA-256").Msg("Configuring Kafka SASL SCRAM-SHA-256 authentication")
		case "SCRAM-SHA-512":
			saslOpt = kgo.SASL(scram.Auth{
				User: cfg.SASL.Username,
				Pass: cfg.SASL.Password,
			}.AsSha512Mechanism())
			logger.Info().Str("mechanism", "SCRAM-SHA-512").Msg("Configuring Kafka SASL SCRAM-SHA-512 authentication")
		default:
			logger.Warn().Str("mechanism", cfg.SASL.Mechanism).Msg("Unsupported Kafka SASL mechanism")
			return nil, fmt.Errorf("NewProducer: unsupported Kafka SASL mechanism: %s", cfg.SASL.Mechanism)
		}
		opts = append(opts, saslOpt)
	} else {
		logger.Info().Msg("Kafka SASL authentication is disabled.")
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to create franz-go Kafka client")
		return nil, fmt.Errorf("NewProducer: failed to create franz-go Kafka client: %w", err)
	}

	// Ping the Kafka cluster to ensure connectivity on startup.
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer pingCancel()
	if err := client.Ping(pingCtx); err != nil {
		client.Close() // Important to close client if setup fails partway
		logger.Warn().Err(err).Strs("brokers", cfg.Brokers).Msg("Failed to ping Kafka cluster")
		return nil, fmt.Errorf("NewProducer: failed to ping Kafka cluster with brokers %v: %w", cfg.Brokers, err)
	}

	logger.Info().Strs("brokers", cfg.Brokers).Msg("Successfully connected and pinged Kafka brokers")

	return &franzKafkaProducer{
		client:      client,
		logger:      logger.With().Str("module", "kafka_producer").Logger(),
		topicPrefix: cfg.TopicPrefix,
	}, nil
}

// Produce sends a message to Kafka.
// The topic provided will be prefixed with the configured TopicPrefix if it's not empty.
func (p *franzKafkaProducer) Produce(ctx context.Context, topic string, key, value []byte) error {
	fullTopic := topic
	if p.topicPrefix != "" {
		// Ensure no double underscores or awkward joins if prefix already ends with separator
		// or topic starts with one. This is a simple concatenation.
		fullTopic = p.topicPrefix + topic
	}

	record := &kgo.Record{
		Topic: fullTopic,
		Key:   key,
		Value: value,
		// Timestamp: time.Now(), // franz-go sets this by default if not specified.
		// Headers can be added here if needed:
		// Headers: []kgo.RecordHeader{{Key: "my-header", Value: []byte("my-value")}},
	}

	p.logger.Debug().Str("topic", fullTopic).Int("key_len", len(key)).Int("value_len", len(value)).Msg("Producing message to Kafka")

	const maxAttempts = 3
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		_, lastErr = p.client.ProduceSync(ctx, record).First()
		if lastErr == nil {
			p.logger.Debug().Str("topic", fullTopic).Msg("Successfully produced message to Kafka")
			return nil
		}
		p.logger.Warn().Err(lastErr).Str("topic", fullTopic).Int("attempt", attempt).Int("max_attempts", maxAttempts).Msg("Failed to produce message to Kafka")
		if attempt < maxAttempts {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * 500 * time.Millisecond):
			}
		}
	}
	return fmt.Errorf("franzKafkaProducer.Produce: failed to produce to topic %s after %d attempts: %w", fullTopic, maxAttempts, lastErr)
}

// Close gracefully shuts down the Kafka client.
// This will attempt to flush any buffered messages.
func (p *franzKafkaProducer) Close() error {
	p.logger.Info().Msg("Closing Kafka producer client...")
	// kgo.Client.Close() is a blocking call that waits for buffered messages to be flushed
	// or for the context passed to Produce calls to expire.
	p.client.Close() // franz-go's Close() itself doesn't return an error.
	p.logger.Info().Msg("Kafka producer client closed.")
	return nil
}

// wait or retry kafka logic currently using scripts to handle it
func waitForKafka(log zerolog.Logger, brokers []string, retries int, delay time.Duration) {
	waitLogger := log.With().Str("component", "wait_for_kafka").Logger()
	for i := 0; i < retries; i++ {
		for _, broker := range brokers {
			host, port, err := net.SplitHostPort(broker)
			if err != nil {
				waitLogger.Warn().
					Err(err).
					Str("broker", broker).
					Msg("Invalid broker format while waiting for Kafka")
				continue
			}

			conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 2*time.Second)
			if err == nil {
				conn.Close()
				waitLogger.Info().
					Str("broker", broker).
					Msg("Kafka broker is reachable")
				return
			}

			waitLogger.Warn().
				Err(err).
				Str("broker", broker).
				Msg("Kafka broker not reachable yet")
		}
		waitLogger.Warn().
			Int("attempt", i+1).
			Int("retries", retries).
			Dur("delay", delay).
			Msg("Kafka not reachable yet, retrying after delay")
		time.Sleep(delay)
	}

	waitLogger.Error().Msg("Kafka is not reachable after retries. Exiting.")
	os.Exit(1)
}
