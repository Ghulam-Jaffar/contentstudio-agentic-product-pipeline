package main

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ServiceConfig holds configuration for the analytics sink service
type ServiceConfig struct {
	PostsParserWorkers     int
	InsightsParserWorkers  int
	BatchProcessorsPerType int
	MaxBatchSize           int
	BatchTimeout           time.Duration
	IdleTimeout            time.Duration
	ParseChanSize          int
	MessageChanSize        int
}

// DefaultServiceConfig returns the default service configuration
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		PostsParserWorkers:     2,
		InsightsParserWorkers:  2,
		BatchProcessorsPerType: batchProcessorsPerType,
		MaxBatchSize:           maxBatchSize,
		BatchTimeout:           batchTimeout,
		IdleTimeout:            idleTimeout,
		ParseChanSize:          1000,
		MessageChanSize:        messageChanSize,
	}
}

// ServiceDependencies holds all external dependencies
type ServiceDependencies struct {
	Sink             ClickHouseSinkInterface
	PostsConsumer    KafkaConsumerInterface
	InsightsConsumer KafkaConsumerInterface
	Logger           *logger.Logger
}

// ServiceMetrics holds runtime metrics
type ServiceMetrics struct {
	PostsMessagesReceived    uint64
	InsightsMessagesReceived uint64
	PostsMessagesParsed      uint64
	InsightsMessagesParsed   uint64
	MessagesBatched          uint64
	MessagesFailed           uint64
	InsertErrors             uint64
}

// RunService starts the TikTok analytics sink service
func RunService(ctx context.Context, deps *ServiceDependencies, cfg ServiceConfig) error {
	log := deps.Logger

	metrics := &ServiceMetrics{}

	log.Info().
		Int("posts_parser_workers", cfg.PostsParserWorkers).
		Int("insights_parser_workers", cfg.InsightsParserWorkers).
		Int("max_batch_size", cfg.MaxBatchSize).
		Dur("batch_timeout", cfg.BatchTimeout).
		Msg("Starting TikTok Analytics Sink")

	var wg sync.WaitGroup

	// Posts parser and ClickHouse sink
	wg.Add(1)
	go processPosts(ctx, deps, &wg, cfg, metrics, log)

	// Insights parser and ClickHouse sink
	wg.Add(1)
	go processInsights(ctx, deps, &wg, cfg, metrics, log)

	wg.Wait()

	log.Info().
		Uint64("posts_received", atomic.LoadUint64(&metrics.PostsMessagesReceived)).
		Uint64("insights_received", atomic.LoadUint64(&metrics.InsightsMessagesReceived)).
		Uint64("parsed_total", atomic.LoadUint64(&metrics.PostsMessagesParsed)+atomic.LoadUint64(&metrics.InsightsMessagesParsed)).
		Uint64("insert_errors", atomic.LoadUint64(&metrics.InsertErrors)).
		Msg("TikTok Analytics Sink stopped")

	return nil
}

// processPosts handles TikTok posts: parse and insert to ClickHouse
func processPosts(ctx context.Context, deps *ServiceDependencies, wg *sync.WaitGroup, cfg ServiceConfig, metrics *ServiceMetrics, log *logger.Logger) {
	defer wg.Done()

	postsLog := log.Logger.With().Str("type", "posts").Logger()

	batch := make([]*clickhousemodels.TikTokPosts, 0, cfg.MaxBatchSize)
	batchTimer := time.NewTimer(cfg.BatchTimeout)
	defer batchTimer.Stop()

	err := deps.PostsConsumer.Consume(ctx, []string{rawPostsTopic}, func(ctx context.Context, topic string, key, value []byte) error {
		var rawPost kafkamodels.RawTikTokPost
		if err := json.Unmarshal(value, &rawPost); err != nil {
			postsLog.Error().Err(err).Msg("Failed to unmarshal post")
			atomic.AddUint64(&metrics.MessagesFailed, 1)
			return nil
		}

		atomic.AddUint64(&metrics.PostsMessagesReceived, 1)

		// Parse post
		parsedPost := parsePost(&rawPost)
		if parsedPost == nil {
			postsLog.Warn().Str("tiktok_id", rawPost.TikTokID).Msg("Failed to parse post")
			return nil
		}

		atomic.AddUint64(&metrics.PostsMessagesParsed, 1)

		batch = append(batch, parsedPost)

		// Flush if batch is full
		if len(batch) >= cfg.MaxBatchSize {
			if err := deps.Sink.BulkInsertTikTokPosts(ctx, batch); err != nil {
				postsLog.Error().Err(err).Int("batch_size", len(batch)).Msg("Failed to insert batch")
				atomic.AddUint64(&metrics.InsertErrors, 1)
			} else {
				atomic.AddUint64(&metrics.MessagesBatched, uint64(len(batch)))
			}
			batch = make([]*clickhousemodels.TikTokPosts, 0, cfg.MaxBatchSize)
			batchTimer.Reset(cfg.BatchTimeout)
		}

		return nil
	})

	if err != nil && err != context.Canceled {
		postsLog.Error().Err(err).Msg("Posts consumer error")
	}
}

// processInsights handles TikTok insights: parse and insert to ClickHouse
func processInsights(ctx context.Context, deps *ServiceDependencies, wg *sync.WaitGroup, cfg ServiceConfig, metrics *ServiceMetrics, log *logger.Logger) {
	defer wg.Done()

	insightsLog := log.Logger.With().Str("type", "insights").Logger()

	batch := make([]*clickhousemodels.TikTokInsights, 0, cfg.MaxBatchSize)
	batchTimer := time.NewTimer(cfg.BatchTimeout)
	defer batchTimer.Stop()

	err := deps.InsightsConsumer.Consume(ctx, []string{rawInsightsTopic}, func(ctx context.Context, topic string, key, value []byte) error {
		var rawInsight kafkamodels.ParsedTikTokInsights
		if err := json.Unmarshal(value, &rawInsight); err != nil {
			insightsLog.Error().Err(err).Msg("Failed to unmarshal insight")
			atomic.AddUint64(&metrics.MessagesFailed, 1)
			return nil
		}

		atomic.AddUint64(&metrics.InsightsMessagesReceived, 1)

		// Parse insight
		parsedInsight := parseInsight(&rawInsight)
		if parsedInsight == nil {
			insightsLog.Warn().Str("record_id", rawInsight.RecordID).Msg("Failed to parse insight")
			return nil
		}

		atomic.AddUint64(&metrics.InsightsMessagesParsed, 1)

		batch = append(batch, parsedInsight)

		// Flush if batch is full
		if len(batch) >= cfg.MaxBatchSize {
			if err := deps.Sink.BulkInsertTikTokInsights(ctx, batch); err != nil {
				insightsLog.Error().Err(err).Int("batch_size", len(batch)).Msg("Failed to insert batch")
				atomic.AddUint64(&metrics.InsertErrors, 1)
			} else {
				atomic.AddUint64(&metrics.MessagesBatched, uint64(len(batch)))
			}
			batch = make([]*clickhousemodels.TikTokInsights, 0, cfg.MaxBatchSize)
			batchTimer.Reset(cfg.BatchTimeout)
		}

		return nil
	})

	if err != nil && err != context.Canceled {
		insightsLog.Error().Err(err).Msg("Insights consumer error")
	}
}

// parsePost converts raw post to parsed post for ClickHouse
// Note: In actual implementation, use parsing.NewTikTokParser().ParseVideo()
func parsePost(raw *kafkamodels.RawTikTokPost) *clickhousemodels.TikTokPosts {
	if raw == nil {
		return nil
	}

	// This is a simplified placeholder - actual parsing happens in main.go via parser.ParseVideo()
	return &clickhousemodels.TikTokPosts{
		TikTokID: raw.TikTokID,
	}
}

// parseInsight converts parsed insight to ClickHouse model
func parseInsight(raw *kafkamodels.ParsedTikTokInsights) *clickhousemodels.TikTokInsights {
	if raw == nil {
		return nil
	}

	return &clickhousemodels.TikTokInsights{
		RecordID:            raw.RecordID,
		TikTokID:            raw.TikTokID,
		DisplayName:         raw.DisplayName,
		ProfileImage:        raw.ProfileImage,
		TotalFollowerCount:  raw.TotalFollowerCount,
		TotalFollowingCount: raw.TotalFollowingCount,
		TotalLikeCount:      raw.TotalLikeCount,
		TotalVideoCount:     raw.TotalVideoCount,
		TotalVideoViews:     raw.TotalVideoViews,
		TotalVideoLikes:     raw.TotalVideoLikes,
		TotalVideoComments:  raw.TotalVideoComments,
		TotalVideoShares:    raw.TotalVideoShares,
		IsVerified:          raw.IsVerified,
		Bio:                 raw.Bio,
		ProfileLink:         raw.ProfileLink,
	}
}

// GetMetrics returns current service metrics
func GetMetrics(metrics *ServiceMetrics) map[string]uint64 {
	return map[string]uint64{
		"posts_received":    atomic.LoadUint64(&metrics.PostsMessagesReceived),
		"insights_received": atomic.LoadUint64(&metrics.InsightsMessagesReceived),
		"posts_parsed":      atomic.LoadUint64(&metrics.PostsMessagesParsed),
		"insights_parsed":   atomic.LoadUint64(&metrics.InsightsMessagesParsed),
		"messages_batched":  atomic.LoadUint64(&metrics.MessagesBatched),
		"messages_failed":   atomic.LoadUint64(&metrics.MessagesFailed),
		"insert_errors":     atomic.LoadUint64(&metrics.InsertErrors),
	}
}
