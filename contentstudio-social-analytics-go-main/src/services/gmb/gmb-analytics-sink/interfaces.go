package main

import (
	"context"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

// ClickHouseSinkInterface defines the interface for GMB-specific ClickHouse operations.
type ClickHouseSinkInterface interface {
	BulkInsertGMBDailyMetrics(ctx context.Context, metrics []*clickhousemodels.GMBDailyMetrics) error
	BulkInsertGMBMediaAssets(ctx context.Context, assets []*clickhousemodels.GMBMediaAssets) error
	BulkInsertGMBSearchKeywordsMonthly(ctx context.Context, keywords []*clickhousemodels.GMBSearchKeywordsMonthly) error
	BulkInsertGMBLocalPosts(ctx context.Context, posts []*clickhousemodels.GMBLocalPosts) error
	BulkInsertGMBReviews(ctx context.Context, reviews []*clickhousemodels.GMBReviews) error
}

// KafkaConsumer is an alias to the shared interface in kafka package.
type KafkaConsumer = kafka.Consumer

// KafkaConsumerInterface is an alias for backward compatibility.
type KafkaConsumerInterface = kafka.Consumer

// MessageHandler is an alias to the shared type in kafka package.
type MessageHandler = kafka.MessageHandler
