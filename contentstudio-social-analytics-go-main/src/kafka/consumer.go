package kafka

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"
	"github.com/twmb/franz-go/plugin/kzerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
)

// MessageHandler is a function type for handling consumed messages.
// It receives the context, topic name, message key, and message value.
type MessageHandler func(ctx context.Context, topic string, key, value []byte) error

// AcknowledgingMessageHandler is like MessageHandler but receives an ack function.
// The ack function MUST be called exactly once when processing is fully complete
// (whether success or permanent failure). Calling ack marks the record's offset
// ready to be committed; until ack is called the offset is held back so the
// message will be redelivered on restart.
type AcknowledgingMessageHandler func(ctx context.Context, topic string, key, value []byte, ack func()) error

// Consumer defines the interface for a Kafka consumer.
type Consumer interface {
	// Consume starts consuming messages from the specified topics.
	// The handler function will be called for each message.
	// This is a blocking operation that runs until the context is cancelled.
	Consume(ctx context.Context, topics []string, handler MessageHandler) error
	// ConsumeWithAck is like Consume but provides at-least-once delivery guarantees.
	// Offsets are committed only after the handler calls ack(), so in-flight jobs
	// are redelivered on restart. The ack function is safe to call from any goroutine.
	ConsumeWithAck(ctx context.Context, topics []string, handler AcknowledgingMessageHandler) error
	// Close shuts down the consumer and releases resources.
	Close() error
}

type franzKafkaConsumer struct {
	client      *kgo.Client
	logger      zerolog.Logger
	topicPrefix string
	groupID     string
	cfg         config.KafkaConfig
}

// NewConsumer creates a new Kafka consumer instance using franz-go.
// It takes Kafka configuration, consumer group ID, and a Zerolog logger.
func NewConsumer(cfg config.KafkaConfig, groupID string, logger zerolog.Logger) (Consumer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, errors.New("kafka brokers must be configured")
	}

	if groupID == "" {
		return nil, errors.New("consumer group ID must be provided")
	}

	return &franzKafkaConsumer{
		logger:      logger.With().Str("module", "kafka_consumer").Str("group_id", groupID).Logger(),
		topicPrefix: cfg.TopicPrefix,
		groupID:     groupID,
		cfg:         cfg,
	}, nil
}

// Consume starts consuming messages from the specified topics.
// The topics will be prefixed with the configured TopicPrefix if it's not empty.
func (c *franzKafkaConsumer) Consume(ctx context.Context, topics []string, handler MessageHandler) error {
	if len(topics) == 0 {
		return errors.New("at least one topic must be specified")
	}

	// Apply topic prefix to all topics
	fullTopics := make([]string, len(topics))
	for i, topic := range topics {
		fullTopic := topic
		if c.topicPrefix != "" {
			fullTopic = c.topicPrefix + topic
		}
		fullTopics[i] = fullTopic
	}

	c.logger.Info().Strs("topics", fullTopics).Msg("Starting to consume from Kafka topics")

	// Create client with topics
	kgoClientLogger := c.logger.With().Str("component", "kgo_client").Logger()
	opts := []kgo.Opt{
		kgo.SeedBrokers(c.cfg.Brokers...),
		kgo.WithLogger(kzerolog.New(&kgoClientLogger)),
		// Configure consumer group
		kgo.ConsumerGroup(c.groupID),
		// Configure topics to consume
		kgo.ConsumeTopics(fullTopics...),
		// Configure session timeout and heartbeat
		kgo.SessionTimeout(20 * time.Second),
		kgo.HeartbeatInterval(3 * time.Second),
		// Start consuming from the beginning for new consumer groups to process all existing messages
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
		// Configure fetch settings
		kgo.FetchMaxWait(5 * time.Second),
		kgo.FetchMinBytes(1),
		kgo.FetchMaxBytes(50 * 1024 * 1024), // 50MB max fetch
	}

	// Add SASL authentication if configured
	if c.cfg.SASL.Enabled {
		if c.cfg.SASL.Username == "" || c.cfg.SASL.Password == "" {
			c.logger.Warn().Msg("Kafka SASL is enabled but username or password is not configured")
			return errors.New("kafka SASL enabled but username or password missing")
		}
		var saslOpt kgo.Opt
		switch c.cfg.SASL.Mechanism {
		case "PLAIN":
			saslOpt = kgo.SASL(plain.Auth{
				User: c.cfg.SASL.Username,
				Pass: c.cfg.SASL.Password,
			}.AsMechanism())
			c.logger.Info().Str("mechanism", "PLAIN").Msg("Configuring Kafka SASL PLAIN authentication")
		case "SCRAM-SHA-256":
			saslOpt = kgo.SASL(scram.Auth{
				User: c.cfg.SASL.Username,
				Pass: c.cfg.SASL.Password,
			}.AsSha256Mechanism())
			c.logger.Info().Str("mechanism", "SCRAM-SHA-256").Msg("Configuring Kafka SASL SCRAM-SHA-256 authentication")
		case "SCRAM-SHA-512":
			saslOpt = kgo.SASL(scram.Auth{
				User: c.cfg.SASL.Username,
				Pass: c.cfg.SASL.Password,
			}.AsSha512Mechanism())
			c.logger.Info().Str("mechanism", "SCRAM-SHA-512").Msg("Configuring Kafka SASL SCRAM-SHA-512 authentication")
		default:
			c.logger.Warn().Str("mechanism", c.cfg.SASL.Mechanism).Msg("Unsupported Kafka SASL mechanism")
			return fmt.Errorf("franzKafkaConsumer.Consume: unsupported Kafka SASL mechanism: %s", c.cfg.SASL.Mechanism)
		}
		opts = append(opts, saslOpt)
	} else {
		c.logger.Info().Msg("Kafka SASL authentication is disabled.")
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		c.logger.Warn().Err(err).Msg("Failed to create franz-go Kafka client")
		return fmt.Errorf("franzKafkaConsumer.Consume: failed to create franz-go Kafka client: %w", err)
	}
	c.client = client
	defer c.client.Close()

	// Ping the Kafka cluster to ensure connectivity on startup.
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer pingCancel()
	if err := c.client.Ping(pingCtx); err != nil {
		c.logger.Warn().Err(err).Strs("brokers", c.cfg.Brokers).Msg("Failed to ping Kafka cluster")
		return fmt.Errorf("franzKafkaConsumer.Consume: failed to ping Kafka cluster with brokers %v: %w", c.cfg.Brokers, err)
	}

	c.logger.Info().Strs("brokers", c.cfg.Brokers).Str("group_id", c.groupID).Msg("Successfully connected and pinged Kafka brokers")

	// Start consuming messages
	for {
		select {
		case <-ctx.Done():
			c.logger.Info().Msg("Context cancelled, stopping consumer")
			return ctx.Err()
		default:
			// Poll for messages
			fetches := c.client.PollFetches(ctx)
			if errs := fetches.Errors(); len(errs) > 0 {
				for _, err := range errs {
					if errors.Is(err.Err, context.Canceled) || errors.Is(err.Err, context.DeadlineExceeded) {
						continue
					}
					c.logger.Warn().Err(err.Err).Str("topic", err.Topic).Int32("partition", err.Partition).
						Msg("Error encountered while consuming from Kafka")
				}
				if ctx.Err() != nil {
					return ctx.Err()
				}
				continue
			}

			// Process each message
			fetches.EachPartition(func(p kgo.FetchTopicPartition) {
				for _, record := range p.Records {
					c.logger.Debug().
						Str("topic", record.Topic).
						Int32("partition", record.Partition).
						Int64("offset", record.Offset).
						Int("key_len", len(record.Key)).
						Int("value_len", len(record.Value)).
						Msg("Processing message")

					// Call the message handler
					if err := handler(ctx, record.Topic, record.Key, record.Value); err != nil {
						c.logger.Warn().Err(err).
							Str("topic", record.Topic).
							Int32("partition", record.Partition).
							Int64("offset", record.Offset).
							Msg("Error processing message")
						// Note: In a production system, you might want to implement retry logic,
						// dead letter queue, or other error handling strategies here.
						continue
					}

					c.logger.Debug().
						Str("topic", record.Topic).
						Int32("partition", record.Partition).
						Int64("offset", record.Offset).
						Msg("Successfully processed message")
				}
			})

			// Commit offsets
			if err := c.client.CommitUncommittedOffsets(ctx); err != nil {
				c.logger.Warn().Err(err).Msg("Failed to commit offsets")
			}
		}
	}
}

// Close gracefully shuts down the Kafka consumer client.
func (c *franzKafkaConsumer) Close() error {
	c.logger.Info().Msg("Closing Kafka consumer client...")
	if c.client != nil {
		c.client.Close()
	}
	c.logger.Info().Msg("Kafka consumer client closed.")
	return nil
}

// ConsumeWithAck consumes messages and provides at-least-once delivery.
// Offsets are committed only after the handler calls the supplied ack function,
// so jobs that are in-flight when the process restarts are redelivered.
// The ack function uses MarkCommitRecords + CommitMarkedOffsets (not CommitUncommittedOffsets),
// so unacked records are never inadvertently committed.
func (c *franzKafkaConsumer) ConsumeWithAck(ctx context.Context, topics []string, handler AcknowledgingMessageHandler) error {
	if len(topics) == 0 {
		return errors.New("at least one topic must be specified")
	}

	fullTopics := make([]string, len(topics))
	for i, topic := range topics {
		fullTopic := topic
		if c.topicPrefix != "" {
			fullTopic = c.topicPrefix + topic
		}
		fullTopics[i] = fullTopic
	}

	c.logger.Info().Strs("topics", fullTopics).Msg("Starting ConsumeWithAck from Kafka topics")

	kgoClientLogger := c.logger.With().Str("component", "kgo_client").Logger()
	opts := []kgo.Opt{
		kgo.SeedBrokers(c.cfg.Brokers...),
		kgo.WithLogger(kzerolog.New(&kgoClientLogger)),
		kgo.ConsumerGroup(c.groupID),
		kgo.ConsumeTopics(fullTopics...),
		kgo.SessionTimeout(20 * time.Second),
		kgo.HeartbeatInterval(3 * time.Second),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
		kgo.FetchMaxWait(5 * time.Second),
		kgo.FetchMinBytes(1),
		kgo.FetchMaxBytes(50 * 1024 * 1024),
		kgo.DisableAutoCommit(),
	}

	if c.cfg.SASL.Enabled {
		if c.cfg.SASL.Username == "" || c.cfg.SASL.Password == "" {
			return errors.New("kafka SASL enabled but username or password missing")
		}
		var saslOpt kgo.Opt
		switch c.cfg.SASL.Mechanism {
		case "PLAIN":
			saslOpt = kgo.SASL(plain.Auth{User: c.cfg.SASL.Username, Pass: c.cfg.SASL.Password}.AsMechanism())
		case "SCRAM-SHA-256":
			saslOpt = kgo.SASL(scram.Auth{User: c.cfg.SASL.Username, Pass: c.cfg.SASL.Password}.AsSha256Mechanism())
		case "SCRAM-SHA-512":
			saslOpt = kgo.SASL(scram.Auth{User: c.cfg.SASL.Username, Pass: c.cfg.SASL.Password}.AsSha512Mechanism())
		default:
			return fmt.Errorf("franzKafkaConsumer.ConsumeWithAck: unsupported Kafka SASL mechanism: %s", c.cfg.SASL.Mechanism)
		}
		opts = append(opts, saslOpt)
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return fmt.Errorf("franzKafkaConsumer.ConsumeWithAck: failed to create client: %w", err)
	}
	defer client.Close()

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer pingCancel()
	if err := client.Ping(pingCtx); err != nil {
		return fmt.Errorf("franzKafkaConsumer.ConsumeWithAck: failed to ping Kafka: %w", err)
	}

	// Background goroutine commits marked offsets every 5 seconds.
	commitDone := make(chan struct{})
	go func() {
		defer close(commitDone)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				// Final commit on shutdown.
				_ = client.CommitMarkedOffsets(context.Background())
				return
			case <-ticker.C:
				if err := client.CommitMarkedOffsets(ctx); err != nil && ctx.Err() == nil {
					c.logger.Warn().Err(err).Msg("ConsumeWithAck: failed to commit marked offsets")
				}
			}
		}
	}()
	defer func() { <-commitDone }()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fetches := client.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			for _, fe := range errs {
				if errors.Is(fe.Err, context.Canceled) || errors.Is(fe.Err, context.DeadlineExceeded) {
					continue
				}
				c.logger.Warn().Err(fe.Err).Str("topic", fe.Topic).Int32("partition", fe.Partition).Msg("ConsumeWithAck: fetch error")
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			continue
		}

		fetches.EachPartition(func(p kgo.FetchTopicPartition) {
			for i := range p.Records {
				rec := p.Records[i]
				ack := func() {
					client.MarkCommitRecords(rec)
				}
				c.logger.Debug().Str("topic", rec.Topic).Int32("partition", rec.Partition).Int64("offset", rec.Offset).Msg("ConsumeWithAck: dispatching record")
				if err := handler(ctx, rec.Topic, rec.Key, rec.Value, ack); err != nil {
					c.logger.Warn().Err(err).Str("topic", rec.Topic).Int32("partition", rec.Partition).Int64("offset", rec.Offset).Msg("ConsumeWithAck: handler error; offset will be redelivered on restart")
				}
			}
		})
	}
}
