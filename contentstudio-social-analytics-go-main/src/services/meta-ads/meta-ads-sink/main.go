package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/telemetry"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

const (
	consumerGroupSuffix = "meta-ads-sink-group"

	topicAccountInfo      = "raw-meta-ads-account-info"
	topicCampaigns        = "raw-meta-ads-campaigns"
	topicAdsets           = "raw-meta-ads-adsets"
	topicAds              = "raw-meta-ads-ads"
	topicCampaignInsights = "raw-meta-ads-campaign-insights"
	topicAdsetInsights    = "raw-meta-ads-adset-insights"
	topicAdInsights       = "raw-meta-ads-ad-insights"
	topicAgeGender        = "raw-meta-ads-demographics-age-gender"
	topicDevicePlatform   = "raw-meta-ads-demographics-device-platform"
	topicRegionCountry    = "raw-meta-ads-demographics-region-country"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting Meta Ads Sink service")

	sink := conversions.NewClickHouseSink(&log.Logger, cfg)
	if err := sink.Health(); err != nil {
		log.Warn().Err(err).Msg("ClickHouse health check failed - continuing anyway")
	}

	topics := []string{
		topicAccountInfo,
		topicCampaigns,
		topicAdsets,
		topicAds,
		topicCampaignInsights,
		topicAdsetInsights,
		topicAdInsights,
		topicAgeGender,
		topicDevicePlatform,
		topicRegionCountry,
	}

	consumer, err := kafka2.NewConsumer(cfg.Kafka, consumerGroupSuffix, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka consumer")
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info().Msg("Shutdown signal received")
		cancel()
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info().Strs("topics", topics).Str("consumer_group", consumerGroupSuffix).Msg("Starting consumer")

		if err := consumer.Consume(ctx, topics,
			func(ctx context.Context, topic string, key, value []byte) error {
				return dispatch(ctx, topic, value, sink, log)
			}); err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("Consumer error")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("Waiting for in-flight messages to complete")
	wg.Wait()
	log.Info().Msg("Meta Ads Sink service stopped")
}

func dispatch(
	ctx context.Context,
	topic string,
	value []byte,
	sink *conversions.ClickHouseSink,
	log *logger.Logger,
) error {
	switch topic {
	case topicAccountInfo:
		var payload kafkamodels.MetaAdsAccountInfoPayload
		if err := json.Unmarshal(value, &payload); err != nil {
			log.Error().Err(err).Str("topic", topic).Msg("Failed to unmarshal payload")
			return nil
		}
		row := sink.ConvertMetaAdsAccountInfo(payload.WorkOrder.PlatformIdentifier, payload.AccountInfo)
		start := time.Now()
		if err := sink.BulkInsertMetaAdsAccountInfo(ctx, []*clickhousemodels.MetaAdsAccountInfo{row}); err != nil {
			log.Error().Err(err).Str("topic", topic).Str("account_id", payload.WorkOrder.PlatformIdentifier).Msg("Failed to insert account info")
		} else {
			log.Info().Int("rows", 1).Str("topic", topic).Str("account_id", payload.WorkOrder.PlatformIdentifier).Dur("duration", time.Since(start)).Msg("Inserted rows into ClickHouse")
		}

	case topicCampaigns:
		var payload kafkamodels.MetaAdsCampaignsPayload
		if err := json.Unmarshal(value, &payload); err != nil {
			log.Error().Err(err).Str("topic", topic).Msg("Failed to unmarshal payload")
			return nil
		}
		rows := make([]*clickhousemodels.MetaAdsCampaign, 0, len(payload.Campaigns))
		for _, r := range payload.Campaigns {
			rows = append(rows, sink.ConvertMetaAdsCampaign(payload.WorkOrder.PlatformIdentifier, r))
		}
		start := time.Now()
		if err := sink.BulkInsertMetaAdsCampaigns(ctx, rows); err != nil {
			log.Error().Err(err).Str("topic", topic).Str("account_id", payload.WorkOrder.PlatformIdentifier).Msg("Failed to insert campaigns")
		} else {
			log.Info().Int("rows", len(rows)).Str("topic", topic).Str("account_id", payload.WorkOrder.PlatformIdentifier).Dur("duration", time.Since(start)).Msg("Inserted rows into ClickHouse")
		}

	case topicAdsets:
		var payload kafkamodels.MetaAdsAdsetsPayload
		if err := json.Unmarshal(value, &payload); err != nil {
			log.Error().Err(err).Str("topic", topic).Msg("Failed to unmarshal payload")
			return nil
		}
		rows := make([]*clickhousemodels.MetaAdsAdset, 0, len(payload.Adsets))
		for _, r := range payload.Adsets {
			rows = append(rows, sink.ConvertMetaAdsAdset(payload.WorkOrder.PlatformIdentifier, r))
		}
		start := time.Now()
		if err := sink.BulkInsertMetaAdsAdsets(ctx, rows); err != nil {
			log.Error().Err(err).Str("topic", topic).Str("account_id", payload.WorkOrder.PlatformIdentifier).Msg("Failed to insert adsets")
		} else {
			log.Info().Int("rows", len(rows)).Str("topic", topic).Str("account_id", payload.WorkOrder.PlatformIdentifier).Dur("duration", time.Since(start)).Msg("Inserted rows into ClickHouse")
		}

	case topicAds:
		var payload kafkamodels.MetaAdsAdsPayload
		if err := json.Unmarshal(value, &payload); err != nil {
			log.Error().Err(err).Str("topic", topic).Msg("Failed to unmarshal payload")
			return nil
		}
		rows := make([]*clickhousemodels.MetaAdsAd, 0, len(payload.Ads))
		for _, r := range payload.Ads {
			rows = append(rows, sink.ConvertMetaAdsAd(payload.WorkOrder.PlatformIdentifier, r))
		}
		start := time.Now()
		if err := sink.BulkInsertMetaAdsAds(ctx, rows); err != nil {
			log.Error().Err(err).Str("topic", topic).Str("account_id", payload.WorkOrder.PlatformIdentifier).Msg("Failed to insert ads")
		} else {
			log.Info().Int("rows", len(rows)).Str("topic", topic).Str("account_id", payload.WorkOrder.PlatformIdentifier).Dur("duration", time.Since(start)).Msg("Inserted rows into ClickHouse")
		}

	case topicCampaignInsights:
		var payload kafkamodels.MetaAdsCampaignInsightsPayload
		if err := json.Unmarshal(value, &payload); err != nil {
			log.Error().Err(err).Str("topic", topic).Msg("Failed to unmarshal payload")
			return nil
		}
		rows := make([]*clickhousemodels.MetaAdsCampaignInsights, 0, len(payload.Insights))
		for _, r := range payload.Insights {
			rows = append(rows, sink.ConvertMetaAdsCampaignInsight(payload.WorkOrder.PlatformIdentifier, r))
		}
		start := time.Now()
		if err := sink.BulkInsertMetaAdsCampaignInsights(ctx, rows); err != nil {
			log.Error().Err(err).Str("topic", topic).Str("account_id", payload.WorkOrder.PlatformIdentifier).Msg("Failed to insert campaign insights")
		} else {
			log.Info().Int("rows", len(rows)).Str("topic", topic).Str("account_id", payload.WorkOrder.PlatformIdentifier).Dur("duration", time.Since(start)).Msg("Inserted rows into ClickHouse")
		}

	case topicAdsetInsights:
		var payload kafkamodels.MetaAdsAdsetInsightsPayload
		if err := json.Unmarshal(value, &payload); err != nil {
			log.Error().Err(err).Str("topic", topic).Msg("Failed to unmarshal payload")
			return nil
		}
		rows := make([]*clickhousemodels.MetaAdsAdsetInsights, 0, len(payload.Insights))
		for _, r := range payload.Insights {
			rows = append(rows, sink.ConvertMetaAdsAdsetInsight(payload.WorkOrder.PlatformIdentifier, r))
		}
		start := time.Now()
		if err := sink.BulkInsertMetaAdsAdsetInsights(ctx, rows); err != nil {
			log.Error().Err(err).Str("topic", topic).Str("account_id", payload.WorkOrder.PlatformIdentifier).Msg("Failed to insert adset insights")
		} else {
			log.Info().Int("rows", len(rows)).Str("topic", topic).Str("account_id", payload.WorkOrder.PlatformIdentifier).Dur("duration", time.Since(start)).Msg("Inserted rows into ClickHouse")
		}

	case topicAdInsights:
		var payload kafkamodels.MetaAdsAdInsightsPayload
		if err := json.Unmarshal(value, &payload); err != nil {
			log.Error().Err(err).Str("topic", topic).Msg("Failed to unmarshal payload")
			return nil
		}
		rows := make([]*clickhousemodels.MetaAdsAdInsights, 0, len(payload.Insights))
		for _, r := range payload.Insights {
			rows = append(rows, sink.ConvertMetaAdsAdInsight(payload.WorkOrder.PlatformIdentifier, r))
		}
		start := time.Now()
		if err := sink.BulkInsertMetaAdsAdInsights(ctx, rows); err != nil {
			log.Error().Err(err).Str("topic", topic).Str("account_id", payload.WorkOrder.PlatformIdentifier).Msg("Failed to insert ad insights")
		} else {
			log.Info().Int("rows", len(rows)).Str("topic", topic).Str("account_id", payload.WorkOrder.PlatformIdentifier).Dur("duration", time.Since(start)).Msg("Inserted rows into ClickHouse")
		}

	case topicAgeGender:
		var payload kafkamodels.MetaAdsDemographicsAgeGenderPayload
		if err := json.Unmarshal(value, &payload); err != nil {
			log.Error().Err(err).Str("topic", topic).Msg("Failed to unmarshal payload")
			return nil
		}
		rows := make([]*clickhousemodels.MetaAdsDemographicsAgeGender, 0, len(payload.Rows))
		for _, r := range payload.Rows {
			rows = append(rows, sink.ConvertMetaAdsDemographicsAgeGender(payload.WorkOrder.PlatformIdentifier, r))
		}
		start := time.Now()
		if err := sink.BulkInsertMetaAdsDemographicsAgeGender(ctx, rows); err != nil {
			log.Error().Err(err).Str("topic", topic).Str("account_id", payload.WorkOrder.PlatformIdentifier).Msg("Failed to insert age/gender demographics")
		} else {
			log.Info().Int("rows", len(rows)).Str("topic", topic).Str("account_id", payload.WorkOrder.PlatformIdentifier).Dur("duration", time.Since(start)).Msg("Inserted rows into ClickHouse")
		}

	case topicDevicePlatform:
		var payload kafkamodels.MetaAdsDemographicsDevicePlatformPayload
		if err := json.Unmarshal(value, &payload); err != nil {
			log.Error().Err(err).Str("topic", topic).Msg("Failed to unmarshal payload")
			return nil
		}
		rows := make([]*clickhousemodels.MetaAdsDemographicsDevicePlatform, 0, len(payload.Rows))
		for _, r := range payload.Rows {
			rows = append(rows, sink.ConvertMetaAdsDemographicsDevicePlatform(payload.WorkOrder.PlatformIdentifier, r))
		}
		start := time.Now()
		if err := sink.BulkInsertMetaAdsDemographicsDevicePlatform(ctx, rows); err != nil {
			log.Error().Err(err).Str("topic", topic).Str("account_id", payload.WorkOrder.PlatformIdentifier).Msg("Failed to insert device/platform demographics")
		} else {
			log.Info().Int("rows", len(rows)).Str("topic", topic).Str("account_id", payload.WorkOrder.PlatformIdentifier).Dur("duration", time.Since(start)).Msg("Inserted rows into ClickHouse")
		}

	case topicRegionCountry:
		var payload kafkamodels.MetaAdsDemographicsRegionCountryPayload
		if err := json.Unmarshal(value, &payload); err != nil {
			log.Error().Err(err).Str("topic", topic).Msg("Failed to unmarshal payload")
			return nil
		}
		rows := make([]*clickhousemodels.MetaAdsDemographicsRegionCountry, 0, len(payload.Rows))
		for _, r := range payload.Rows {
			rows = append(rows, sink.ConvertMetaAdsDemographicsRegionCountry(payload.WorkOrder.PlatformIdentifier, r))
		}
		start := time.Now()
		if err := sink.BulkInsertMetaAdsDemographicsRegionCountry(ctx, rows); err != nil {
			log.Error().Err(err).Str("topic", topic).Str("account_id", payload.WorkOrder.PlatformIdentifier).Msg("Failed to insert region/country demographics")
		} else {
			log.Info().Int("rows", len(rows)).Str("topic", topic).Str("account_id", payload.WorkOrder.PlatformIdentifier).Dur("duration", time.Since(start)).Msg("Inserted rows into ClickHouse")
		}

	default:
		log.Warn().Str("topic", topic).Msg("Unknown topic received by meta-ads-sink, skipping")
	}
	return nil
}
