package processor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/sync/errgroup"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/crypto"
)

const (
	// metaAdsImmediateMaxDateRange is the look-back period for the immediate processor.
	metaAdsImmediateMaxDateRange = 365 * 24 * time.Hour
	// metaAdsBatchSize is the maximum number of rows per ClickHouse batch.
	metaAdsBatchSize = 500
)

// Processor handles the immediate processing of a Meta Ads account.
type Processor struct {
	MongoRepo    mongodb.UnifiedSocialRepository
	MetaAdsAPI   MetaAdsAPIInterface
	Sink         ClickHouseSinkInterface
	Notifier     NotifierInterface
	PusherClient PusherClientInterface
	Logger       *logger.Logger
	Config       *config.Config
}

// New creates a new Processor with all dependencies wired.
func New(
	mongoRepo mongodb.UnifiedSocialRepository,
	sink *conversions.ClickHouseSink,
	notifier NotifierInterface,
	pusherClient PusherClientInterface,
	log *logger.Logger,
	cfg *config.Config,
) *Processor {
	return &Processor{
		MongoRepo:    mongoRepo,
		MetaAdsAPI:   social.NewMetaAdsClient(cfg.Facebook.AppSecret, log),
		Sink:         sink,
		Notifier:     notifier,
		PusherClient: pusherClient,
		Logger:       log,
		Config:       cfg,
	}
}

// ProcessAccount runs the full data pipeline for one Meta Ads account.
func (p *Processor) ProcessAccount(ctx context.Context, wo kafkamodels.MetaAdsWorkOrder) error {
	accountObjectID, err := primitive.ObjectIDFromHex(wo.MongoID)
	if err != nil {
		return fmt.Errorf("MetaAdsProcessor.ProcessAccount: invalid mongo id %q: %w", wo.MongoID, err)
	}

	account, err := p.MongoRepo.FindByID(ctx, accountObjectID)
	if err != nil {
		return fmt.Errorf("MetaAdsProcessor.ProcessAccount: fetch account from mongo: %w", err)
	}
	if account == nil {
		return fmt.Errorf("MetaAdsProcessor.ProcessAccount: account not found: %s", wo.MongoID)
	}

	// Resolve access token: try long_access_token (decrypt), then access_token (decrypt), then plain.
	accessToken, err := p.resolveAccessToken(wo, account)
	if err != nil {
		return fmt.Errorf("MetaAdsProcessor.ProcessAccount: no valid access token: %w", err)
	}

	// Debug token to check validity before fetching data.
	appAccessToken := p.Config.Facebook.AppID + "|" + p.Config.Facebook.AppSecret
	debugResult, err := p.MetaAdsAPI.DebugToken(ctx, accessToken, appAccessToken)
	if err != nil {
		// Network or API error — don't touch MongoDB, just bail and let the job retry.
		p.Logger.Warn().
			Err(err).
			Str("account_id", wo.MongoID).
			Str("platform_identifier", wo.PlatformIdentifier).
			Msg("Meta Ads debug token check failed")
		return fmt.Errorf("MetaAdsProcessor.ProcessAccount: debug token check failed for %s: %w", wo.PlatformIdentifier, err)
	}

	if debugResult != nil && !debugResult.Data.IsValid {
		// Token is explicitly invalid — persist the validity flag.
		p.Logger.Warn().
			Str("account_id", wo.MongoID).
			Str("platform_identifier", wo.PlatformIdentifier).
			Msg("Meta Ads token is invalid; marking account validity invalid")

		_ = p.MongoRepo.UpdateValidity(ctx, accountObjectID, mongomodels.ValidityInvalid)

		// For immediate syncs, set state -> error and send a pusher failure notification.
		if wo.SyncType == "immediate" {
			_ = p.MongoRepo.UpdateState(ctx, accountObjectID, mongomodels.StateFailed)
			if p.PusherClient != nil {
				channel := "meta_ads-analytics-channel-" + wo.WorkspaceID + "-" + wo.PlatformIdentifier
				event := "syncing-" + wo.WorkspaceID + "-" + wo.PlatformIdentifier
				_ = p.PusherClient.Trigger(channel, event, map[string]string{
					"state":               "Failed",
					"account":             wo.PlatformIdentifier,
					"account_id":          wo.MongoID,
					"platform_identifier": wo.PlatformIdentifier,
					"error_type":          "token_invalid",
				})
			}
		}

		return fmt.Errorf("MetaAdsProcessor.ProcessAccount: token invalid for %s", wo.PlatformIdentifier)
	}

	// Resolve date range (immediate: last 1 year).
	since, until := p.resolveDateRange(wo.StartDate, wo.EndDate)

	sinceStr := since.Format("2006-01-02")
	untilStr := until.Format("2006-01-02")

	originalState := account.State
	_ = p.MongoRepo.UpdateState(ctx, accountObjectID, mongomodels.StateSyncing)

	p.Logger.Info().
		Str("account_id", wo.MongoID).
		Str("platform_identifier", wo.PlatformIdentifier).
		Str("since", sinceStr).
		Str("until", untilStr).
		Msg("Starting Meta Ads immediate data fetch")

	if err := p.fetchAndStore(ctx, wo, account, accessToken, since, until); err != nil {
		_ = p.MongoRepo.UpdateState(ctx, accountObjectID, mongomodels.StateFailed)
		return fmt.Errorf("MetaAdsProcessor.ProcessAccount: fetch+store: %w", err)
	}

	now := time.Now().UTC()
	_ = p.MongoRepo.UpdateValidity(ctx, accountObjectID, mongomodels.ValidityValid)
	_ = p.MongoRepo.UpdateState(ctx, accountObjectID, mongomodels.StateProcessed)
	_ = p.MongoRepo.UpdateAnalyticsTimestamp(ctx, accountObjectID, "analytics", now)

	p.Logger.Info().
		Str("account_id", wo.MongoID).
		Str("platform_identifier", wo.PlatformIdentifier).
		Msg("Meta Ads immediate processing completed successfully")

	// Notifications.
	if p.PusherClient != nil {
		channel := "meta_ads-analytics-channel-" + wo.WorkspaceID + "-" + wo.PlatformIdentifier
		event := "syncing-" + wo.WorkspaceID + "-" + wo.PlatformIdentifier
		_ = p.PusherClient.Trigger(channel, event, map[string]string{
			"state":                     "Processed",
			"account":                   wo.PlatformIdentifier,
			"account_id":                wo.MongoID,
			"platform_identifier":       wo.PlatformIdentifier,
			"last_analytics_updated_at": now.Format("2006-01-02"),
		})
	}
	if originalState == mongomodels.StateAdded && p.Notifier != nil {
		_ = p.Notifier.SendAnalyticsNotification(
			account.GetUserIDHex(),
			wo.WorkspaceID,
			mongomodels.PlatformMetaAds,
			wo.MongoID,
			account.PlatformName,
			false,
		)
	}

	return nil
}

// fetchAndStore fetches all 10 endpoints concurrently and stores them in ClickHouse.
func (p *Processor) fetchAndStore(
	ctx context.Context,
	wo kafkamodels.MetaAdsWorkOrder,
	account *mongomodels.SocialIntegration,
	accessToken string,
	since, until time.Time,
) error {
	accountID := wo.PlatformIdentifier // "act_XXXX"
	sinceStr := since.Format("2006-01-02")
	untilStr := until.Format("2006-01-02")
	_ = sinceStr
	_ = untilStr

	eg, egCtx := errgroup.WithContext(ctx)

	// 1. Account info
	eg.Go(func() error {
		raw, err := p.MetaAdsAPI.FetchAccountInfo(egCtx, accountID, accessToken)
		if err != nil {
			p.Logger.Warn().Err(err).Str("endpoint", "account_info").Str("account_id", accountID).Msg("Failed to fetch account info")
			return nil
		}
		row := p.Sink.ConvertMetaAdsAccountInfo(accountID, *raw)
		if err := p.Sink.BulkInsertMetaAdsAccountInfo(egCtx, []*clickhousemodels.MetaAdsAccountInfo{row}); err != nil {
			p.Logger.Error().Err(err).Str("endpoint", "account_info").Msg("Failed to insert account info")
		}
		p.Logger.Info().Str("endpoint", "account_info").Str("account_id", accountID).Msg("Account info stored")
		return nil
	})

	// 2. Campaigns
	eg.Go(func() error {
		rows, err := p.MetaAdsAPI.FetchCampaigns(egCtx, accountID, accessToken, since, until)
		if err != nil {
			p.Logger.Warn().Err(err).Str("endpoint", "campaigns").Str("account_id", accountID).Msg("Failed to fetch campaigns")
			return nil
		}
		converted := make([]*clickhousemodels.MetaAdsCampaign, 0, len(rows))
		for _, r := range rows {
			converted = append(converted, p.Sink.ConvertMetaAdsCampaign(accountID, r))
		}
		batchInsert(egCtx, converted, metaAdsBatchSize, func(batch []*clickhousemodels.MetaAdsCampaign) error {
			return p.Sink.BulkInsertMetaAdsCampaigns(egCtx, batch)
		}, p.Logger, "campaigns")
		p.Logger.Info().Str("endpoint", "campaigns").Int("count", len(rows)).Str("account_id", accountID).Msg("Campaigns stored")
		return nil
	})

	// 3. Ad Sets
	eg.Go(func() error {
		rows, err := p.MetaAdsAPI.FetchAdsets(egCtx, accountID, accessToken, since, until)
		if err != nil {
			p.Logger.Warn().Err(err).Str("endpoint", "adsets").Str("account_id", accountID).Msg("Failed to fetch adsets")
			return nil
		}
		converted := make([]*clickhousemodels.MetaAdsAdset, 0, len(rows))
		for _, r := range rows {
			converted = append(converted, p.Sink.ConvertMetaAdsAdset(accountID, r))
		}
		batchInsert(egCtx, converted, metaAdsBatchSize, func(batch []*clickhousemodels.MetaAdsAdset) error {
			return p.Sink.BulkInsertMetaAdsAdsets(egCtx, batch)
		}, p.Logger, "adsets")
		p.Logger.Info().Str("endpoint", "adsets").Int("count", len(rows)).Str("account_id", accountID).Msg("Adsets stored")
		return nil
	})

	// 4. Ads
	eg.Go(func() error {
		rows, err := p.MetaAdsAPI.FetchAds(egCtx, accountID, accessToken, since, until)
		if err != nil {
			p.Logger.Warn().Err(err).Str("endpoint", "ads").Str("account_id", accountID).Msg("Failed to fetch ads")
			return nil
		}
		converted := make([]*clickhousemodels.MetaAdsAd, 0, len(rows))
		for _, r := range rows {
			converted = append(converted, p.Sink.ConvertMetaAdsAd(accountID, r))
		}
		batchInsert(egCtx, converted, metaAdsBatchSize, func(batch []*clickhousemodels.MetaAdsAd) error {
			return p.Sink.BulkInsertMetaAdsAds(egCtx, batch)
		}, p.Logger, "ads")
		p.Logger.Info().Str("endpoint", "ads").Int("count", len(rows)).Str("account_id", accountID).Msg("Ads stored")
		return nil
	})

	// 5. Campaign insights
	eg.Go(func() error {
		rows, err := p.MetaAdsAPI.FetchCampaignInsights(egCtx, accountID, accessToken, since, until)
		if err != nil {
			p.Logger.Warn().Err(err).Str("endpoint", "campaign_insights").Str("account_id", accountID).Msg("Failed to fetch campaign insights")
			return nil
		}
		converted := make([]*clickhousemodels.MetaAdsCampaignInsights, 0, len(rows))
		for _, r := range rows {
			converted = append(converted, p.Sink.ConvertMetaAdsCampaignInsight(accountID, r))
		}
		batchInsert(egCtx, converted, metaAdsBatchSize, func(batch []*clickhousemodels.MetaAdsCampaignInsights) error {
			return p.Sink.BulkInsertMetaAdsCampaignInsights(egCtx, batch)
		}, p.Logger, "campaign_insights")
		p.Logger.Info().Str("endpoint", "campaign_insights").Int("count", len(rows)).Str("account_id", accountID).Msg("Campaign insights stored")
		return nil
	})

	// 6. Adset insights
	eg.Go(func() error {
		rows, err := p.MetaAdsAPI.FetchAdsetInsights(egCtx, accountID, accessToken, since, until)
		if err != nil {
			p.Logger.Warn().Err(err).Str("endpoint", "adset_insights").Str("account_id", accountID).Msg("Failed to fetch adset insights")
			return nil
		}
		converted := make([]*clickhousemodels.MetaAdsAdsetInsights, 0, len(rows))
		for _, r := range rows {
			converted = append(converted, p.Sink.ConvertMetaAdsAdsetInsight(accountID, r))
		}
		batchInsert(egCtx, converted, metaAdsBatchSize, func(batch []*clickhousemodels.MetaAdsAdsetInsights) error {
			return p.Sink.BulkInsertMetaAdsAdsetInsights(egCtx, batch)
		}, p.Logger, "adset_insights")
		p.Logger.Info().Str("endpoint", "adset_insights").Int("count", len(rows)).Str("account_id", accountID).Msg("Adset insights stored")
		return nil
	})

	// 7. Ad insights
	eg.Go(func() error {
		rows, err := p.MetaAdsAPI.FetchAdInsights(egCtx, accountID, accessToken, since, until)
		if err != nil {
			p.Logger.Warn().Err(err).Str("endpoint", "ad_insights").Str("account_id", accountID).Msg("Failed to fetch ad insights")
			return nil
		}
		converted := make([]*clickhousemodels.MetaAdsAdInsights, 0, len(rows))
		for _, r := range rows {
			converted = append(converted, p.Sink.ConvertMetaAdsAdInsight(accountID, r))
		}
		batchInsert(egCtx, converted, metaAdsBatchSize, func(batch []*clickhousemodels.MetaAdsAdInsights) error {
			return p.Sink.BulkInsertMetaAdsAdInsights(egCtx, batch)
		}, p.Logger, "ad_insights")
		p.Logger.Info().Str("endpoint", "ad_insights").Int("count", len(rows)).Str("account_id", accountID).Msg("Ad insights stored")
		return nil
	})

	// 8. Demographics age/gender
	eg.Go(func() error {
		rows, err := p.MetaAdsAPI.FetchAgeGenderInsights(egCtx, accountID, accessToken, since, until)
		if err != nil {
			p.Logger.Warn().Err(err).Str("endpoint", "demographics_age_gender").Str("account_id", accountID).Msg("Failed to fetch age/gender demographics")
			return nil
		}
		converted := make([]*clickhousemodels.MetaAdsDemographicsAgeGender, 0, len(rows))
		for _, r := range rows {
			converted = append(converted, p.Sink.ConvertMetaAdsDemographicsAgeGender(accountID, r))
		}
		batchInsert(egCtx, converted, metaAdsBatchSize, func(batch []*clickhousemodels.MetaAdsDemographicsAgeGender) error {
			return p.Sink.BulkInsertMetaAdsDemographicsAgeGender(egCtx, batch)
		}, p.Logger, "demographics_age_gender")
		p.Logger.Info().Str("endpoint", "demographics_age_gender").Int("count", len(rows)).Str("account_id", accountID).Msg("Age/gender demographics stored")
		return nil
	})

	// 9. Demographics device/platform
	eg.Go(func() error {
		rows, err := p.MetaAdsAPI.FetchDevicePlatformInsights(egCtx, accountID, accessToken, since, until)
		if err != nil {
			p.Logger.Warn().Err(err).Str("endpoint", "demographics_device_platform").Str("account_id", accountID).Msg("Failed to fetch device/platform demographics")
			return nil
		}
		converted := make([]*clickhousemodels.MetaAdsDemographicsDevicePlatform, 0, len(rows))
		for _, r := range rows {
			converted = append(converted, p.Sink.ConvertMetaAdsDemographicsDevicePlatform(accountID, r))
		}
		batchInsert(egCtx, converted, metaAdsBatchSize, func(batch []*clickhousemodels.MetaAdsDemographicsDevicePlatform) error {
			return p.Sink.BulkInsertMetaAdsDemographicsDevicePlatform(egCtx, batch)
		}, p.Logger, "demographics_device_platform")
		p.Logger.Info().Str("endpoint", "demographics_device_platform").Int("count", len(rows)).Str("account_id", accountID).Msg("Device/platform demographics stored")
		return nil
	})

	// 10. Demographics region/country
	eg.Go(func() error {
		rows, err := p.MetaAdsAPI.FetchRegionCountryInsights(egCtx, accountID, accessToken, since, until)
		if err != nil {
			p.Logger.Warn().Err(err).Str("endpoint", "demographics_region_country").Str("account_id", accountID).Msg("Failed to fetch region/country demographics")
			return nil
		}
		converted := make([]*clickhousemodels.MetaAdsDemographicsRegionCountry, 0, len(rows))
		for _, r := range rows {
			converted = append(converted, p.Sink.ConvertMetaAdsDemographicsRegionCountry(accountID, r))
		}
		batchInsert(egCtx, converted, metaAdsBatchSize, func(batch []*clickhousemodels.MetaAdsDemographicsRegionCountry) error {
			return p.Sink.BulkInsertMetaAdsDemographicsRegionCountry(egCtx, batch)
		}, p.Logger, "demographics_region_country")
		p.Logger.Info().Str("endpoint", "demographics_region_country").Int("count", len(rows)).Str("account_id", accountID).Msg("Region/country demographics stored")
		return nil
	})

	return eg.Wait()
}

// resolveAccessToken tries long_access_token (decrypted), then access_token (decrypted), then plain.
func (p *Processor) resolveAccessToken(wo kafkamodels.MetaAdsWorkOrder, account *mongomodels.SocialIntegration) (string, error) {
	if wo.LongAccessToken != "" {
		if dec, err := crypto.DecryptToken(wo.LongAccessToken, p.Config.DecryptionKey); err == nil && dec != "" {
			return dec, nil
		}
		return wo.LongAccessToken, nil
	}
	longToken := account.GetAccessToken()
	if longToken != "" {
		if dec, err := crypto.DecryptToken(longToken, p.Config.DecryptionKey); err == nil && dec != "" {
			return dec, nil
		}
		return longToken, nil
	}
	if wo.AccessToken != "" {
		if dec, err := crypto.DecryptToken(wo.AccessToken, p.Config.DecryptionKey); err == nil && dec != "" {
			return dec, nil
		}
		return wo.AccessToken, nil
	}
	return "", fmt.Errorf("no access token available")
}

// resolveDateRange returns the date range to use: if start/end are provided use them,
// otherwise default to last 1 year.
func (p *Processor) resolveDateRange(startDate, endDate string) (time.Time, time.Time) {
	now := time.Now().UTC()
	defaultUntil := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)
	defaultSince := defaultUntil.Add(-metaAdsImmediateMaxDateRange)

	startDate = strings.TrimSpace(startDate)
	endDate = strings.TrimSpace(endDate)
	if startDate != "" && endDate != "" {
		if s, err := time.Parse("2006-01-02", startDate); err == nil {
			if e, err := time.Parse("2006-01-02", endDate); err == nil {
				return s.UTC(), e.UTC()
			}
		}
	}
	return defaultSince, defaultUntil
}

// batchInsert splits rows into chunks of size batchSize and calls insertFn for each chunk.
func batchInsert[T any](
	ctx context.Context,
	rows []T,
	batchSize int,
	insertFn func([]T) error,
	log *logger.Logger,
	endpoint string,
) {
	for i := 0; i < len(rows); i += batchSize {
		end := i + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		if err := insertFn(rows[i:end]); err != nil {
			log.Error().Err(err).Str("endpoint", endpoint).Int("batch_start", i).Int("batch_end", end).Msg("Failed to insert batch")
		}
	}
}
