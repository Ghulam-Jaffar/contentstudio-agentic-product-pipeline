package conversions

import (
	"context"
	"strconv"
	"strings"
	"time"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func parseMetaFloat64(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}

func parseMetaInt64(s string) int64 {
	v, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return v
}

func parseMetaDate(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", strings.TrimSpace(s))
	if err != nil {
		return time.Time{}
	}
	return t.UTC()
}

func findActionValue(actions []kafkamodels.RawMetaAdsAction, actionType string) int64 {
	for _, a := range actions {
		if a.ActionType == actionType {
			return parseMetaInt64(a.Value)
		}
	}
	return 0
}

// ─────────────────────────────────────────────────────────────────────────────
// Account Info
// ─────────────────────────────────────────────────────────────────────────────

// ConvertMetaAdsAccountInfo converts raw API account info to the ClickHouse model.
func (s *ClickHouseSink) ConvertMetaAdsAccountInfo(accountID string, raw kafkamodels.RawMetaAdsAccountInfo) *clickhousemodels.MetaAdsAccountInfo {
	row := &clickhousemodels.MetaAdsAccountInfo{
		AccountID:     accountID,
		Name:          raw.Name,
		Currency:      raw.Currency,
		AccountStatus: raw.AccountStatus,
		TimezoneName:  raw.TimezoneName,
		AmountSpent:   raw.AmountSpent,
		Balance:       raw.Balance,
		SpendCap:      raw.SpendCap,
		CreatedTime:   raw.CreatedTime.Time,
		InsertedAt:    truncateToHourUTC(time.Now()),
	}
	if raw.Business != nil {
		row.BusinessID = raw.Business.ID
		row.BusinessName = raw.Business.Name
	}
	return row
}

// BulkInsertMetaAdsAccountInfo converts and inserts account info via the sink.
func (s *ClickHouseSink) BulkInsertMetaAdsAccountInfo(ctx context.Context, rows []*clickhousemodels.MetaAdsAccountInfo) error {
	if len(rows) == 0 {
		return nil
	}
	s.logger.Info().Int("count", len(rows)).Msg("Bulk inserting Meta Ads account info to ClickHouse")
	return s.ClickhouseClient.BulkInsertMetaAdsAccountInfo(ctx, rows)
}

// ─────────────────────────────────────────────────────────────────────────────
// Campaigns
// ─────────────────────────────────────────────────────────────────────────────

// ConvertMetaAdsCampaign converts a raw API campaign to the ClickHouse model.
func (s *ClickHouseSink) ConvertMetaAdsCampaign(accountID string, raw kafkamodels.RawMetaAdsCampaign) *clickhousemodels.MetaAdsCampaign {
	return &clickhousemodels.MetaAdsCampaign{
		AccountID:       accountID,
		CampaignID:      raw.ID,
		Name:            raw.Name,
		Status:          raw.Status,
		EffectiveStatus: raw.EffectiveStatus,
		Objective:       raw.Objective,
		DailyBudget:     raw.DailyBudget,
		LifetimeBudget:  raw.LifetimeBudget,
		BudgetRemaining: raw.BudgetRemaining,
		StartTime:       raw.StartTime.Time,
		StopTime:        raw.StopTime.Time,
		CreatedTime:     raw.CreatedTime.Time,
		UpdatedTime:     raw.UpdatedTime.Time,
		InsertedAt:      truncateToHourUTC(time.Now()),
	}
}

// BulkInsertMetaAdsCampaigns converts and inserts campaign rows.
func (s *ClickHouseSink) BulkInsertMetaAdsCampaigns(ctx context.Context, rows []*clickhousemodels.MetaAdsCampaign) error {
	if len(rows) == 0 {
		return nil
	}
	s.logger.Info().Int("count", len(rows)).Msg("Bulk inserting Meta Ads campaigns to ClickHouse")
	return s.ClickhouseClient.BulkInsertMetaAdsCampaigns(ctx, rows)
}

// ─────────────────────────────────────────────────────────────────────────────
// Ad Sets
// ─────────────────────────────────────────────────────────────────────────────

// ConvertMetaAdsAdset converts a raw API ad set to the ClickHouse model.
func (s *ClickHouseSink) ConvertMetaAdsAdset(accountID string, raw kafkamodels.RawMetaAdsAdset) *clickhousemodels.MetaAdsAdset {
	row := &clickhousemodels.MetaAdsAdset{
		AccountID:        accountID,
		AdsetID:          raw.ID,
		Name:             raw.Name,
		CampaignID:       raw.CampaignID,
		Status:           raw.Status,
		EffectiveStatus:  raw.EffectiveStatus,
		DailyBudget:      raw.DailyBudget,
		LifetimeBudget:   raw.LifetimeBudget,
		BudgetRemaining:  raw.BudgetRemaining,
		BillingEvent:     raw.BillingEvent,
		OptimizationGoal: raw.OptimizationGoal,
		BidStrategy:      raw.BidStrategy,
		TargetingJSON:    string(raw.RawTargeting),
		StartTime:        raw.StartTime.Time,
		StopTime:         raw.StopTime.Time,
		EndTime:          raw.EndTime.Time,
		CreatedTime:      raw.CreatedTime.Time,
		InsertedAt:       truncateToHourUTC(time.Now()),
	}
	if raw.Targeting != nil {
		row.AgeMin = raw.Targeting.AgeMin
		row.AgeMax = raw.Targeting.AgeMax
		if raw.Targeting.GeoLocations != nil {
			row.TargetingCountries = raw.Targeting.GeoLocations.Countries
		}
	}
	if row.TargetingCountries == nil {
		row.TargetingCountries = []string{}
	}
	return row
}

// BulkInsertMetaAdsAdsets converts and inserts ad set rows.
func (s *ClickHouseSink) BulkInsertMetaAdsAdsets(ctx context.Context, rows []*clickhousemodels.MetaAdsAdset) error {
	if len(rows) == 0 {
		return nil
	}
	s.logger.Info().Int("count", len(rows)).Msg("Bulk inserting Meta Ads adsets to ClickHouse")
	return s.ClickhouseClient.BulkInsertMetaAdsAdsets(ctx, rows)
}

// ─────────────────────────────────────────────────────────────────────────────
// Ads
// ─────────────────────────────────────────────────────────────────────────────

// ConvertMetaAdsAd converts a raw API ad to the ClickHouse model.
func (s *ClickHouseSink) ConvertMetaAdsAd(accountID string, raw kafkamodels.RawMetaAdsAd) *clickhousemodels.MetaAdsAd {
	row := &clickhousemodels.MetaAdsAd{
		AccountID:       accountID,
		AdID:            raw.ID,
		Name:            raw.Name,
		AdsetID:         raw.AdsetID,
		CampaignID:      raw.CampaignID,
		Status:          raw.Status,
		EffectiveStatus: raw.EffectiveStatus,
		DailyBudget:     raw.DailyBudget,
		LifetimeBudget:  raw.LifetimeBudget,
		BudgetRemaining: raw.BudgetRemaining,
		CreatedTime:     raw.CreatedTime.Time,
		UpdatedTime:     raw.UpdatedTime.Time,
		InsertedAt:      truncateToHourUTC(time.Now()),
	}
	if raw.Adset != nil {
		row.AdsetName = raw.Adset.Name
	}
	if raw.Campaign != nil {
		row.CampaignName = raw.Campaign.Name
		row.Objective = raw.Campaign.Objective
	}
	if raw.Creative != nil {
		row.CreativeID = raw.Creative.ID
		row.CreativeName = raw.Creative.Name
		row.CreativeTitle = raw.Creative.Title
		row.CreativeBody = raw.Creative.Body
		row.CreativeImageURL = raw.Creative.ImageURL
		row.CreativeThumbnailURL = raw.Creative.ThumbnailURL
		row.CreativeObjectType = raw.Creative.ObjectType
		row.CreativeEffectiveObjectStoryID = raw.Creative.EffectiveObjectStoryID
	}
	return row
}

// BulkInsertMetaAdsAds converts and inserts ad rows.
func (s *ClickHouseSink) BulkInsertMetaAdsAds(ctx context.Context, rows []*clickhousemodels.MetaAdsAd) error {
	if len(rows) == 0 {
		return nil
	}
	s.logger.Info().Int("count", len(rows)).Msg("Bulk inserting Meta Ads ads to ClickHouse")
	return s.ClickhouseClient.BulkInsertMetaAdsAds(ctx, rows)
}

// ─────────────────────────────────────────────────────────────────────────────
// Campaign Insights
// ─────────────────────────────────────────────────────────────────────────────

// ConvertMetaAdsCampaignInsight converts one raw insight row (campaign level) to ClickHouse model.
func (s *ClickHouseSink) ConvertMetaAdsCampaignInsight(accountID string, raw kafkamodels.RawMetaAdsInsightRow) *clickhousemodels.MetaAdsCampaignInsights {
	return &clickhousemodels.MetaAdsCampaignInsights{
		AccountID:                               accountID,
		CampaignID:                              raw.CampaignID,
		CampaignName:                            raw.CampaignName,
		Objective:                               raw.Objective,
		InsightsDate:                            parseMetaDate(raw.DateStart),
		Spend:                                   parseMetaFloat64(raw.Spend),
		Impressions:                             parseMetaInt64(raw.Impressions),
		Reach:                                   parseMetaInt64(raw.Reach),
		Clicks:                                  parseMetaInt64(raw.Clicks),
		UniqueClicks:                            parseMetaInt64(raw.UniqueClicks),
		CTR:                                     parseMetaFloat64(raw.CTR),
		UniqueCTR:                               parseMetaFloat64(raw.UniqueCTR),
		CPC:                                     parseMetaFloat64(raw.CPC),
		CPM:                                     parseMetaFloat64(raw.CPM),
		CPP:                                     parseMetaFloat64(raw.CPP),
		Frequency:                               parseMetaFloat64(raw.Frequency),
		ActionsPurchase:                         findActionValue(raw.Actions, "purchase"),
		ActionsPostEngagement:                   findActionValue(raw.Actions, "post_engagement"),
		ActionsOffsiteConversionFbPixelPurchase: findActionValue(raw.Actions, "offsite_conversion.fb_pixel_purchase"),
		ActionsLinkClick:                        findActionValue(raw.Actions, "link_click"),
		ActionsLead:                             findActionValue(raw.Actions, "lead"),
		ActionsOffsiteConversionFbPixelLead:     findActionValue(raw.Actions, "offsite_conversion.fb_pixel_lead"),
		ActionsMobileAppInstall:                 findActionValue(raw.Actions, "mobile_app_install"),
		InsertedAt:                              truncateToHourUTC(time.Now()),
	}
}

// BulkInsertMetaAdsCampaignInsights converts and inserts campaign insight rows.
func (s *ClickHouseSink) BulkInsertMetaAdsCampaignInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsCampaignInsights) error {
	if len(rows) == 0 {
		return nil
	}
	s.logger.Info().Int("count", len(rows)).Msg("Bulk inserting Meta Ads campaign insights to ClickHouse")
	return s.ClickhouseClient.BulkInsertMetaAdsCampaignInsights(ctx, rows)
}

// ─────────────────────────────────────────────────────────────────────────────
// Adset Insights
// ─────────────────────────────────────────────────────────────────────────────

// ConvertMetaAdsAdsetInsight converts one raw insight row (adset level) to ClickHouse model.
func (s *ClickHouseSink) ConvertMetaAdsAdsetInsight(accountID string, raw kafkamodels.RawMetaAdsInsightRow) *clickhousemodels.MetaAdsAdsetInsights {
	return &clickhousemodels.MetaAdsAdsetInsights{
		AccountID:                               accountID,
		AdsetID:                                 raw.AdsetID,
		AdsetName:                               raw.AdsetName,
		CampaignID:                              raw.CampaignID,
		CampaignName:                            raw.CampaignName,
		InsightsDate:                            parseMetaDate(raw.DateStart),
		Spend:                                   parseMetaFloat64(raw.Spend),
		Impressions:                             parseMetaInt64(raw.Impressions),
		Reach:                                   parseMetaInt64(raw.Reach),
		Clicks:                                  parseMetaInt64(raw.Clicks),
		UniqueClicks:                            parseMetaInt64(raw.UniqueClicks),
		CTR:                                     parseMetaFloat64(raw.CTR),
		UniqueCTR:                               parseMetaFloat64(raw.UniqueCTR),
		CPC:                                     parseMetaFloat64(raw.CPC),
		CPM:                                     parseMetaFloat64(raw.CPM),
		CPP:                                     parseMetaFloat64(raw.CPP),
		Frequency:                               parseMetaFloat64(raw.Frequency),
		ActionsPurchase:                         findActionValue(raw.Actions, "purchase"),
		ActionsPostEngagement:                   findActionValue(raw.Actions, "post_engagement"),
		ActionsOffsiteConversionFbPixelPurchase: findActionValue(raw.Actions, "offsite_conversion.fb_pixel_purchase"),
		ActionsLinkClick:                        findActionValue(raw.Actions, "link_click"),
		ActionsLead:                             findActionValue(raw.Actions, "lead"),
		ActionsOffsiteConversionFbPixelLead:     findActionValue(raw.Actions, "offsite_conversion.fb_pixel_lead"),
		ActionsMobileAppInstall:                 findActionValue(raw.Actions, "mobile_app_install"),
		InsertedAt:                              truncateToHourUTC(time.Now()),
	}
}

// BulkInsertMetaAdsAdsetInsights converts and inserts adset insight rows.
func (s *ClickHouseSink) BulkInsertMetaAdsAdsetInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsAdsetInsights) error {
	if len(rows) == 0 {
		return nil
	}
	s.logger.Info().Int("count", len(rows)).Msg("Bulk inserting Meta Ads adset insights to ClickHouse")
	return s.ClickhouseClient.BulkInsertMetaAdsAdsetInsights(ctx, rows)
}

// ─────────────────────────────────────────────────────────────────────────────
// Ad Insights
// ─────────────────────────────────────────────────────────────────────────────

// ConvertMetaAdsAdInsight converts one raw insight row (ad level) to ClickHouse model.
func (s *ClickHouseSink) ConvertMetaAdsAdInsight(accountID string, raw kafkamodels.RawMetaAdsInsightRow) *clickhousemodels.MetaAdsAdInsights {
	return &clickhousemodels.MetaAdsAdInsights{
		AccountID:                               accountID,
		AdID:                                    raw.AdID,
		AdName:                                  raw.AdName,
		AdsetID:                                 raw.AdsetID,
		CampaignID:                              raw.CampaignID,
		CampaignName:                            raw.CampaignName,
		InsightsDate:                            parseMetaDate(raw.DateStart),
		Spend:                                   parseMetaFloat64(raw.Spend),
		Impressions:                             parseMetaInt64(raw.Impressions),
		Reach:                                   parseMetaInt64(raw.Reach),
		Clicks:                                  parseMetaInt64(raw.Clicks),
		UniqueClicks:                            parseMetaInt64(raw.UniqueClicks),
		CTR:                                     parseMetaFloat64(raw.CTR),
		UniqueCTR:                               parseMetaFloat64(raw.UniqueCTR),
		CPC:                                     parseMetaFloat64(raw.CPC),
		CPM:                                     parseMetaFloat64(raw.CPM),
		CPP:                                     parseMetaFloat64(raw.CPP),
		Frequency:                               parseMetaFloat64(raw.Frequency),
		ActionsPurchase:                         findActionValue(raw.Actions, "purchase"),
		ActionsPostEngagement:                   findActionValue(raw.Actions, "post_engagement"),
		ActionsOffsiteConversionFbPixelPurchase: findActionValue(raw.Actions, "offsite_conversion.fb_pixel_purchase"),
		ActionsLinkClick:                        findActionValue(raw.Actions, "link_click"),
		ActionsLead:                             findActionValue(raw.Actions, "lead"),
		ActionsOffsiteConversionFbPixelLead:     findActionValue(raw.Actions, "offsite_conversion.fb_pixel_lead"),
		ActionsMobileAppInstall:                 findActionValue(raw.Actions, "mobile_app_install"),
		InsertedAt:                              truncateToHourUTC(time.Now()),
	}
}

// BulkInsertMetaAdsAdInsights converts and inserts ad insight rows.
func (s *ClickHouseSink) BulkInsertMetaAdsAdInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsAdInsights) error {
	if len(rows) == 0 {
		return nil
	}
	s.logger.Info().Int("count", len(rows)).Msg("Bulk inserting Meta Ads ad insights to ClickHouse")
	return s.ClickhouseClient.BulkInsertMetaAdsAdInsights(ctx, rows)
}

// ─────────────────────────────────────────────────────────────────────────────
// Demographics: Age & Gender
// ─────────────────────────────────────────────────────────────────────────────

// ConvertMetaAdsDemographicsAgeGender converts a raw demographics row (age/gender breakdown).
func (s *ClickHouseSink) ConvertMetaAdsDemographicsAgeGender(accountID string, raw kafkamodels.RawMetaAdsDemographicsRow) *clickhousemodels.MetaAdsDemographicsAgeGender {
	return &clickhousemodels.MetaAdsDemographicsAgeGender{
		AccountID:    accountID,
		InsightsDate: parseMetaDate(raw.DateStart),
		Age:          raw.Age,
		Gender:       raw.Gender,
		Impressions:  parseMetaInt64(raw.Impressions),
		Reach:        parseMetaInt64(raw.Reach),
		Clicks:       parseMetaInt64(raw.Clicks),
		Spend:        parseMetaFloat64(raw.Spend),
		CTR:          parseMetaFloat64(raw.CTR),
		CPM:          parseMetaFloat64(raw.CPM),
		CPC:          parseMetaFloat64(raw.CPC),
		CPP:          parseMetaFloat64(raw.CPP),
		Frequency:    parseMetaFloat64(raw.Frequency),
		InsertedAt:   truncateToHourUTC(time.Now()),
	}
}

// BulkInsertMetaAdsDemographicsAgeGender inserts age/gender breakdown rows.
func (s *ClickHouseSink) BulkInsertMetaAdsDemographicsAgeGender(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsAgeGender) error {
	if len(rows) == 0 {
		return nil
	}
	s.logger.Info().Int("count", len(rows)).Msg("Bulk inserting Meta Ads age/gender demographics to ClickHouse")
	return s.ClickhouseClient.BulkInsertMetaAdsDemographicsAgeGender(ctx, rows)
}

// ─────────────────────────────────────────────────────────────────────────────
// Demographics: Device & Platform
// ─────────────────────────────────────────────────────────────────────────────

// ConvertMetaAdsDemographicsDevicePlatform converts a raw demographics row (device/platform breakdown).
func (s *ClickHouseSink) ConvertMetaAdsDemographicsDevicePlatform(accountID string, raw kafkamodels.RawMetaAdsDemographicsRow) *clickhousemodels.MetaAdsDemographicsDevicePlatform {
	return &clickhousemodels.MetaAdsDemographicsDevicePlatform{
		AccountID:         accountID,
		InsightsDate:      parseMetaDate(raw.DateStart),
		ImpressionDevice:  raw.ImpressionDevice,
		PublisherPlatform: raw.PublisherPlatform,
		PlatformPosition:  raw.PlatformPosition,
		Impressions:       parseMetaInt64(raw.Impressions),
		Reach:             parseMetaInt64(raw.Reach),
		Clicks:            parseMetaInt64(raw.Clicks),
		Spend:             parseMetaFloat64(raw.Spend),
		CTR:               parseMetaFloat64(raw.CTR),
		CPM:               parseMetaFloat64(raw.CPM),
		CPC:               parseMetaFloat64(raw.CPC),
		CPP:               parseMetaFloat64(raw.CPP),
		Frequency:         parseMetaFloat64(raw.Frequency),
		InsertedAt:        truncateToHourUTC(time.Now()),
	}
}

// BulkInsertMetaAdsDemographicsDevicePlatform inserts device/platform breakdown rows.
func (s *ClickHouseSink) BulkInsertMetaAdsDemographicsDevicePlatform(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsDevicePlatform) error {
	if len(rows) == 0 {
		return nil
	}
	s.logger.Info().Int("count", len(rows)).Msg("Bulk inserting Meta Ads device/platform demographics to ClickHouse")
	return s.ClickhouseClient.BulkInsertMetaAdsDemographicsDevicePlatform(ctx, rows)
}

// ─────────────────────────────────────────────────────────────────────────────
// Demographics: Region & Country
// ─────────────────────────────────────────────────────────────────────────────

// ConvertMetaAdsDemographicsRegionCountry converts a raw demographics row (region/country breakdown).
func (s *ClickHouseSink) ConvertMetaAdsDemographicsRegionCountry(accountID string, raw kafkamodels.RawMetaAdsDemographicsRow) *clickhousemodels.MetaAdsDemographicsRegionCountry {
	return &clickhousemodels.MetaAdsDemographicsRegionCountry{
		AccountID:    accountID,
		InsightsDate: parseMetaDate(raw.DateStart),
		Country:      raw.Country,
		Region:       raw.Region,
		Impressions:  parseMetaInt64(raw.Impressions),
		Reach:        parseMetaInt64(raw.Reach),
		Clicks:       parseMetaInt64(raw.Clicks),
		Spend:        parseMetaFloat64(raw.Spend),
		CTR:          parseMetaFloat64(raw.CTR),
		CPM:          parseMetaFloat64(raw.CPM),
		CPC:          parseMetaFloat64(raw.CPC),
		CPP:          parseMetaFloat64(raw.CPP),
		Frequency:    parseMetaFloat64(raw.Frequency),
		InsertedAt:   truncateToHourUTC(time.Now()),
	}
}

// BulkInsertMetaAdsDemographicsRegionCountry inserts region/country breakdown rows.
func (s *ClickHouseSink) BulkInsertMetaAdsDemographicsRegionCountry(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsRegionCountry) error {
	if len(rows) == 0 {
		return nil
	}
	s.logger.Info().Int("count", len(rows)).Msg("Bulk inserting Meta Ads region/country demographics to ClickHouse")
	return s.ClickhouseClient.BulkInsertMetaAdsDemographicsRegionCountry(ctx, rows)
}

// ─────────────────────────────────────────────────────────────────────────────
// Internal helper (local to this file to avoid redeclaration with db/clickhouse pkg)
// ─────────────────────────────────────────────────────────────────────────────

func truncateToHourUTC(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, time.UTC)
}
