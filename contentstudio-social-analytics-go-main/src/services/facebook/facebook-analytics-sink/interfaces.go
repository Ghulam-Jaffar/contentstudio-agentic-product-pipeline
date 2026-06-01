package main

import (
	"context"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ClickHouseSinkInterface defines the interface for Facebook-specific ClickHouse operations.
// This interface combines conversion and storage methods specific to the Facebook analytics sink.
type ClickHouseSinkInterface interface {
	Health() error
	ConvertFacebookPost(p *kafkamodels.ParsedFacebookPost) *clickhousemodels.FacebookPosts
	ConvertFacebookMediaAssets(a *kafkamodels.ParsedFacebookMediaAsset) *clickhousemodels.FacebookMediaAssets
	ConvertFacebookInsights(ins *kafkamodels.ParsedFacebookInsights) *clickhousemodels.FacebookInsights
	ConvertFacebookVideoInsights(vi *kafkamodels.ParsedFacebookVideoInsights) *clickhousemodels.FacebookVideoInsights
	ConvertFacebookReelsInsights(ri *kafkamodels.ParsedFacebookReelsInsights) *clickhousemodels.FacebookReelsInsights
	BulkInsertPosts(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error
	BulkInsertMediaAssets(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error
	BulkInsertInsights(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error
	BulkInsertVideoInsights(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error
	BulkInsertReelsInsights(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error
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
