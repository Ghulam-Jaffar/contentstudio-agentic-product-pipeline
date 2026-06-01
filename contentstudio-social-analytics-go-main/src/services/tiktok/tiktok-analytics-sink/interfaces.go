package main

import (
	"context"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

// ClickHouseSinkInterface defines the interface for TikTok-specific ClickHouse operations.
// This interface combines storage methods specific to the TikTok analytics sink.
type ClickHouseSinkInterface interface {
	BulkInsertTikTokPosts(ctx context.Context, posts []*clickhousemodels.TikTokPosts) error
	BulkInsertTikTokInsights(ctx context.Context, insights []*clickhousemodels.TikTokInsights) error
	Close() error
}

// KafkaConsumer is an alias to the shared interface in kafka package.
// This allows services to use the common interface for Kafka consumer operations.
type KafkaConsumer = kafka.Consumer

// KafkaConsumerInterface is an alias for backward compatibility.
// Prefer using KafkaConsumer for new code.
type KafkaConsumerInterface = kafka.Consumer

// MessageHandler is an alias to the shared type in kafka package.
// This allows services to use the common message handler type.
type MessageHandler = kafka.MessageHandler
