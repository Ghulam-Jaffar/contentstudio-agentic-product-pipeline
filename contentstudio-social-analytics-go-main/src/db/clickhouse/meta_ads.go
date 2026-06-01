package clickhouse

import (
	"context"
	"fmt"
	"time"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

// truncateToHour returns t truncated to the start of its hour (UTC).
// All Meta Ads datetime fields stored in ClickHouse are truncated to the hour
// so that ReplacingMergeTree(inserted_at) deduplication works correctly.
func truncateToHour(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, time.UTC)
}

// BulkInsertMetaAdsAccountInfo inserts Meta Ads account info rows into ClickHouse.
func (c *Client) BulkInsertMetaAdsAccountInfo(ctx context.Context, rows []*clickhousemodels.MetaAdsAccountInfo) error {
	if len(rows) == 0 {
		return nil
	}
	c.Logger.Info().Str("table", "meta_ads_account_info").Int("count", len(rows)).Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO meta_ads_account_info (
			account_id, name, currency, account_status, timezone_name,
			business_id, business_name, amount_spent, balance, spend_cap,
			created_time, inserted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("BulkInsertMetaAdsAccountInfo: prepare: %w", err)
	}
	now := truncateToHour(time.Now().UTC())
	for _, r := range rows {
		if err := batch.Append(
			r.AccountID, r.Name, r.Currency, r.AccountStatus, r.TimezoneName,
			r.BusinessID, r.BusinessName, r.AmountSpent, r.Balance, r.SpendCap,
			truncateToHour(r.CreatedTime), now,
		); err != nil {
			return fmt.Errorf("BulkInsertMetaAdsAccountInfo: append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("BulkInsertMetaAdsAccountInfo: send: %w", err)
	}
	c.Logger.Info().Str("table", "meta_ads_account_info").Int("count", len(rows)).Msg("Batch insert completed")
	return nil
}

// BulkInsertMetaAdsCampaigns inserts Meta Ads campaign rows into ClickHouse.
func (c *Client) BulkInsertMetaAdsCampaigns(ctx context.Context, rows []*clickhousemodels.MetaAdsCampaign) error {
	if len(rows) == 0 {
		return nil
	}
	c.Logger.Info().Str("table", "meta_ads_campaigns").Int("count", len(rows)).Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO meta_ads_campaigns (
			account_id, campaign_id, name, status, effective_status, objective,
			daily_budget, lifetime_budget, budget_remaining,
			start_time, stop_time, created_time, updated_time, inserted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("BulkInsertMetaAdsCampaigns: prepare: %w", err)
	}
	now := truncateToHour(time.Now().UTC())
	for _, r := range rows {
		if err := batch.Append(
			r.AccountID, r.CampaignID, r.Name, r.Status, r.EffectiveStatus, r.Objective,
			r.DailyBudget, r.LifetimeBudget, r.BudgetRemaining,
			truncateToHour(r.StartTime), truncateToHour(r.StopTime),
			truncateToHour(r.CreatedTime), truncateToHour(r.UpdatedTime), now,
		); err != nil {
			return fmt.Errorf("BulkInsertMetaAdsCampaigns: append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("BulkInsertMetaAdsCampaigns: send: %w", err)
	}
	c.Logger.Info().Str("table", "meta_ads_campaigns").Int("count", len(rows)).Msg("Batch insert completed")
	return nil
}

// BulkInsertMetaAdsAdsets inserts Meta Ads ad set rows into ClickHouse.
func (c *Client) BulkInsertMetaAdsAdsets(ctx context.Context, rows []*clickhousemodels.MetaAdsAdset) error {
	if len(rows) == 0 {
		return nil
	}
	c.Logger.Info().Str("table", "meta_ads_adsets").Int("count", len(rows)).Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO meta_ads_adsets (
			account_id, adset_id, name, campaign_id, status, effective_status,
			daily_budget, lifetime_budget, budget_remaining,
			billing_event, optimization_goal, bid_strategy,
			age_min, age_max, targeting_countries, targeting_json,
			start_time, stop_time, end_time, created_time, inserted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("BulkInsertMetaAdsAdsets: prepare: %w", err)
	}
	now := truncateToHour(time.Now().UTC())
	for _, r := range rows {
		if err := batch.Append(
			r.AccountID, r.AdsetID, r.Name, r.CampaignID, r.Status, r.EffectiveStatus,
			r.DailyBudget, r.LifetimeBudget, r.BudgetRemaining,
			r.BillingEvent, r.OptimizationGoal, r.BidStrategy,
			r.AgeMin, r.AgeMax, r.TargetingCountries, r.TargetingJSON,
			truncateToHour(r.StartTime), truncateToHour(r.StopTime),
			truncateToHour(r.EndTime), truncateToHour(r.CreatedTime), now,
		); err != nil {
			return fmt.Errorf("BulkInsertMetaAdsAdsets: append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("BulkInsertMetaAdsAdsets: send: %w", err)
	}
	c.Logger.Info().Str("table", "meta_ads_adsets").Int("count", len(rows)).Msg("Batch insert completed")
	return nil
}

// BulkInsertMetaAdsAds inserts Meta Ads ad rows into ClickHouse.
func (c *Client) BulkInsertMetaAdsAds(ctx context.Context, rows []*clickhousemodels.MetaAdsAd) error {
	if len(rows) == 0 {
		return nil
	}
	c.Logger.Info().Str("table", "meta_ads_ads").Int("count", len(rows)).Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO meta_ads_ads (
			account_id, ad_id, name, adset_id, adset_name, campaign_id, campaign_name,
			status, effective_status, objective,
			creative_id, creative_name, creative_title, creative_body,
			creative_image_url, creative_thumbnail_url, creative_object_type, creative_effective_object_story_id,
			daily_budget, lifetime_budget, budget_remaining,
			created_time, updated_time, inserted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("BulkInsertMetaAdsAds: prepare: %w", err)
	}
	now := truncateToHour(time.Now().UTC())
	for _, r := range rows {
		if err := batch.Append(
			r.AccountID, r.AdID, r.Name, r.AdsetID, r.AdsetName, r.CampaignID, r.CampaignName,
			r.Status, r.EffectiveStatus, r.Objective,
			r.CreativeID, r.CreativeName, r.CreativeTitle, r.CreativeBody,
			r.CreativeImageURL, r.CreativeThumbnailURL, r.CreativeObjectType, r.CreativeEffectiveObjectStoryID,
			r.DailyBudget, r.LifetimeBudget, r.BudgetRemaining,
			truncateToHour(r.CreatedTime), truncateToHour(r.UpdatedTime), now,
		); err != nil {
			return fmt.Errorf("BulkInsertMetaAdsAds: append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("BulkInsertMetaAdsAds: send: %w", err)
	}
	c.Logger.Info().Str("table", "meta_ads_ads").Int("count", len(rows)).Msg("Batch insert completed")
	return nil
}

// BulkInsertMetaAdsCampaignInsights inserts campaign-level daily insight rows.
func (c *Client) BulkInsertMetaAdsCampaignInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsCampaignInsights) error {
	if len(rows) == 0 {
		return nil
	}
	c.Logger.Info().Str("table", "meta_ads_campaign_insights").Int("count", len(rows)).Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO meta_ads_campaign_insights (
			account_id, campaign_id, campaign_name, objective, insights_date,
			spend, impressions, reach, clicks, unique_clicks, ctr, unique_ctr, cpc, cpm, cpp, frequency,
			actions_purchase, actions_post_engagement, actions_offsite_conversion_fb_pixel_purchase,
			actions_link_click, actions_lead, actions_offsite_conversion_fb_pixel_lead,
			actions_mobile_app_install,
			inserted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("BulkInsertMetaAdsCampaignInsights: prepare: %w", err)
	}
	now := truncateToHour(time.Now().UTC())
	for _, r := range rows {
		if err := batch.Append(
			r.AccountID, r.CampaignID, r.CampaignName, r.Objective, truncateToHour(r.InsightsDate),
			r.Spend, r.Impressions, r.Reach, r.Clicks, r.UniqueClicks, r.CTR, r.UniqueCTR, r.CPC, r.CPM, r.CPP, r.Frequency,
			r.ActionsPurchase, r.ActionsPostEngagement, r.ActionsOffsiteConversionFbPixelPurchase,
			r.ActionsLinkClick, r.ActionsLead, r.ActionsOffsiteConversionFbPixelLead,
			r.ActionsMobileAppInstall,
			now,
		); err != nil {
			return fmt.Errorf("BulkInsertMetaAdsCampaignInsights: append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("BulkInsertMetaAdsCampaignInsights: send: %w", err)
	}
	c.Logger.Info().Str("table", "meta_ads_campaign_insights").Int("count", len(rows)).Msg("Batch insert completed")
	return nil
}

// BulkInsertMetaAdsAdsetInsights inserts adset-level daily insight rows.
func (c *Client) BulkInsertMetaAdsAdsetInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsAdsetInsights) error {
	if len(rows) == 0 {
		return nil
	}
	c.Logger.Info().Str("table", "meta_ads_adset_insights").Int("count", len(rows)).Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO meta_ads_adset_insights (
			account_id, adset_id, adset_name, campaign_id, campaign_name, insights_date,
			spend, impressions, reach, clicks, unique_clicks, ctr, unique_ctr, cpc, cpm, cpp, frequency,
			actions_purchase, actions_post_engagement, actions_offsite_conversion_fb_pixel_purchase,
			actions_link_click, actions_lead, actions_offsite_conversion_fb_pixel_lead,
			actions_mobile_app_install,
			inserted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("BulkInsertMetaAdsAdsetInsights: prepare: %w", err)
	}
	now := truncateToHour(time.Now().UTC())
	for _, r := range rows {
		if err := batch.Append(
			r.AccountID, r.AdsetID, r.AdsetName, r.CampaignID, r.CampaignName, truncateToHour(r.InsightsDate),
			r.Spend, r.Impressions, r.Reach, r.Clicks, r.UniqueClicks, r.CTR, r.UniqueCTR, r.CPC, r.CPM, r.CPP, r.Frequency,
			r.ActionsPurchase, r.ActionsPostEngagement, r.ActionsOffsiteConversionFbPixelPurchase,
			r.ActionsLinkClick, r.ActionsLead, r.ActionsOffsiteConversionFbPixelLead,
			r.ActionsMobileAppInstall,
			now,
		); err != nil {
			return fmt.Errorf("BulkInsertMetaAdsAdsetInsights: append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("BulkInsertMetaAdsAdsetInsights: send: %w", err)
	}
	c.Logger.Info().Str("table", "meta_ads_adset_insights").Int("count", len(rows)).Msg("Batch insert completed")
	return nil
}

// BulkInsertMetaAdsAdInsights inserts ad-level daily insight rows.
func (c *Client) BulkInsertMetaAdsAdInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsAdInsights) error {
	if len(rows) == 0 {
		return nil
	}
	c.Logger.Info().Str("table", "meta_ads_ad_insights").Int("count", len(rows)).Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO meta_ads_ad_insights (
			account_id, ad_id, ad_name, adset_id, campaign_id, campaign_name, insights_date,
			spend, impressions, reach, clicks, unique_clicks, ctr, unique_ctr, cpc, cpm, cpp, frequency,
			actions_purchase, actions_post_engagement, actions_offsite_conversion_fb_pixel_purchase,
			actions_link_click, actions_lead, actions_offsite_conversion_fb_pixel_lead,
			actions_mobile_app_install,
			inserted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("BulkInsertMetaAdsAdInsights: prepare: %w", err)
	}
	now := truncateToHour(time.Now().UTC())
	for _, r := range rows {
		if err := batch.Append(
			r.AccountID, r.AdID, r.AdName, r.AdsetID, r.CampaignID, r.CampaignName, truncateToHour(r.InsightsDate),
			r.Spend, r.Impressions, r.Reach, r.Clicks, r.UniqueClicks, r.CTR, r.UniqueCTR, r.CPC, r.CPM, r.CPP, r.Frequency,
			r.ActionsPurchase, r.ActionsPostEngagement, r.ActionsOffsiteConversionFbPixelPurchase,
			r.ActionsLinkClick, r.ActionsLead, r.ActionsOffsiteConversionFbPixelLead,
			r.ActionsMobileAppInstall,
			now,
		); err != nil {
			return fmt.Errorf("BulkInsertMetaAdsAdInsights: append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("BulkInsertMetaAdsAdInsights: send: %w", err)
	}
	c.Logger.Info().Str("table", "meta_ads_ad_insights").Int("count", len(rows)).Msg("Batch insert completed")
	return nil
}

// BulkInsertMetaAdsDemographicsAgeGender inserts age/gender breakdown rows.
func (c *Client) BulkInsertMetaAdsDemographicsAgeGender(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsAgeGender) error {
	if len(rows) == 0 {
		return nil
	}
	c.Logger.Info().Str("table", "meta_ads_demographics_age_gender").Int("count", len(rows)).Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO meta_ads_demographics_age_gender (
			account_id, insights_date, age, gender,
			impressions, reach, clicks, spend, ctr, cpm, cpc, cpp, frequency,
			inserted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("BulkInsertMetaAdsDemographicsAgeGender: prepare: %w", err)
	}
	now := truncateToHour(time.Now().UTC())
	for _, r := range rows {
		if err := batch.Append(
			r.AccountID, truncateToHour(r.InsightsDate), r.Age, r.Gender,
			r.Impressions, r.Reach, r.Clicks, r.Spend, r.CTR, r.CPM, r.CPC, r.CPP, r.Frequency,
			now,
		); err != nil {
			return fmt.Errorf("BulkInsertMetaAdsDemographicsAgeGender: append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("BulkInsertMetaAdsDemographicsAgeGender: send: %w", err)
	}
	c.Logger.Info().Str("table", "meta_ads_demographics_age_gender").Int("count", len(rows)).Msg("Batch insert completed")
	return nil
}

// BulkInsertMetaAdsDemographicsDevicePlatform inserts device/platform breakdown rows.
func (c *Client) BulkInsertMetaAdsDemographicsDevicePlatform(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsDevicePlatform) error {
	if len(rows) == 0 {
		return nil
	}
	c.Logger.Info().Str("table", "meta_ads_demographics_device_platform").Int("count", len(rows)).Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO meta_ads_demographics_device_platform (
			account_id, insights_date, impression_device, publisher_platform, platform_position,
			impressions, reach, clicks, spend, ctr, cpm, cpc, cpp, frequency,
			inserted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("BulkInsertMetaAdsDemographicsDevicePlatform: prepare: %w", err)
	}
	now := truncateToHour(time.Now().UTC())
	for _, r := range rows {
		if err := batch.Append(
			r.AccountID, truncateToHour(r.InsightsDate), r.ImpressionDevice, r.PublisherPlatform, r.PlatformPosition,
			r.Impressions, r.Reach, r.Clicks, r.Spend, r.CTR, r.CPM, r.CPC, r.CPP, r.Frequency,
			now,
		); err != nil {
			return fmt.Errorf("BulkInsertMetaAdsDemographicsDevicePlatform: append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("BulkInsertMetaAdsDemographicsDevicePlatform: send: %w", err)
	}
	c.Logger.Info().Str("table", "meta_ads_demographics_device_platform").Int("count", len(rows)).Msg("Batch insert completed")
	return nil
}

// BulkInsertMetaAdsDemographicsRegionCountry inserts region/country breakdown rows.
func (c *Client) BulkInsertMetaAdsDemographicsRegionCountry(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsRegionCountry) error {
	if len(rows) == 0 {
		return nil
	}
	c.Logger.Info().Str("table", "meta_ads_demographics_region_country").Int("count", len(rows)).Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO meta_ads_demographics_region_country (
			account_id, insights_date, country, region,
			impressions, reach, clicks, spend, ctr, cpm, cpc, cpp, frequency,
			inserted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("BulkInsertMetaAdsDemographicsRegionCountry: prepare: %w", err)
	}
	now := truncateToHour(time.Now().UTC())
	for _, r := range rows {
		if err := batch.Append(
			r.AccountID, truncateToHour(r.InsightsDate), r.Country, r.Region,
			r.Impressions, r.Reach, r.Clicks, r.Spend, r.CTR, r.CPM, r.CPC, r.CPP, r.Frequency,
			now,
		); err != nil {
			return fmt.Errorf("BulkInsertMetaAdsDemographicsRegionCountry: append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("BulkInsertMetaAdsDemographicsRegionCountry: send: %w", err)
	}
	c.Logger.Info().Str("table", "meta_ads_demographics_region_country").Int("count", len(rows)).Msg("Batch insert completed")
	return nil
}
